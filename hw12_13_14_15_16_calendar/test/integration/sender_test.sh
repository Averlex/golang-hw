#!/bin/bash

set -e

# --- OS Check ---
# Check if the script is likely running on a native Windows environment
# where it's not intended to run.
#
# Scenarios:
# 1. Running under Cygwin: OSTYPE usually contains "cygwin"
# 2. Running under MSYS2/MINGW: OSTYPE usually contains "msys" or "win"
# 3. Running under WSL: `uname -a` usually contains "microsoft" or "wsl"
# 4. Running under standard Linux/macOS: OSTYPE is like "linux-gnu", "darwin"

# A simple but effective check is to see if OSTYPE contains "win" or "msys"
# or if uname indicates Windows. WSL is generally considered a Linux environment.

if [[ "$OSTYPE" == *"win"* ]] || [[ "$OSTYPE" == *"msys"* ]] || [[ "$OSTYPE" == *"cygwin"* ]]; then
    # Additional check to differentiate WSL from native Windows
    # WSL's `uname -a` typically contains "microsoft" or "wsl"
    if [[ "$(uname -a 2>/dev/null | tr '[:upper:]' '[:lower:]')" == *"microsoft"* ]] || \
       [[ "$(uname -a 2>/dev/null | tr '[:upper:]' '[:lower:]')" == *"wsl"* ]]; then
        # Likely WSL, allow to proceed (you can add a log message if needed)
        # log "Detected WSL, proceeding..."
        :
    else
        # Likely native Windows environment (Cygwin, MSYS2, Git Bash)
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: Running on Windows (non-WSL). This test suite is designed for Linux/macOS/WSL."
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: Skipping notification integration tests."
        # Exit with code 0 to indicate "skipped" rather than "failed"
        exit 0
    fi
fi

# --- End OS Check ---

# --- Configuration ---
CALENDAR_SERVICE_URL="http://localhost:9888/v1/events"
SENDER_CONTAINER_NAME="sender-test"
LOG_LINES_TO_CHECK=100
WAIT_TIME=10
PAST_EVENT_DATETIME="2025-01-01T10:00:00Z"
FUTURE_EVENT_DATETIME="2035-01-01T10:00:00Z"

# --- Functions ---
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to delete an event by ID
# delete_event "event_id"
delete_event() {
    local EVENT_ID_TO_DELETE="$1"
    if [ -z "$EVENT_ID_TO_DELETE" ]; then
        log "WARNING: delete_event called with empty ID. Skipping deletion."
        return 0
    fi

    log "Attempting to delete event with ID: $EVENT_ID_TO_DELETE"
    DELETE_RESPONSE_FILE=$(mktemp)
    DELETE_HTTP_CODE=$(curl -s -S -f -w "%{http_code}" -o "$DELETE_RESPONSE_FILE" -X DELETE \
        "$CALENDAR_SERVICE_URL/$EVENT_ID_TO_DELETE" 2>&1 || echo "curl_error")

    DELETE_RESULT=$?
    if [ $DELETE_RESULT -eq 0 ] && [ "$DELETE_HTTP_CODE" = "200" ]; then
        log "SUCCESS: Event $EVENT_ID_TO_DELETE deleted successfully."
    else
        log "WARNING: Failed to delete event $EVENT_ID_TO_DELETE. HTTP code: $DELETE_HTTP_CODE, curl exit code: $DELETE_RESULT"
        if [ -f "$DELETE_RESPONSE_FILE" ]; then
            log "Delete response body:"
            cat "$DELETE_RESPONSE_FILE"
        fi
    fi
    rm -f "$DELETE_RESPONSE_FILE" # Ensure temp file is always cleaned up
}

# Function to run a single test case
# run_test_case "Test Name" "datetime" "remind_in" "should_find_log"
run_test_case() {
    local TEST_NAME="$1"
    local EVENT_DATETIME="$2"
    local REMIND_IN="$3"
    local SHOULD_FIND_LOG="$4"
    local CREATED_EVENT_ID="" # Variable to hold the ID of the event created in this test

    log ""
    log "=== Running Test Case: $TEST_NAME ==="

    # 1. Define Event Data
    EVENT_DATA=$(cat <<EOF
{
  "title": "$TEST_NAME",
  "datetime": "$EVENT_DATETIME",
  "duration": "3600s",
  "description": "Event for test case: $TEST_NAME",
  "user_id": "integration_test_user",
  "remind_in": "$REMIND_IN"
}
EOF
)

    # 2. Create Event
    log "Creating event..."
    RESPONSE_FILE=$(mktemp)
    HTTP_CODE=$(curl -s -S -f -w "%{http_code}" -o "$RESPONSE_FILE" -X POST \
        -H "Content-Type: application/json" \
        --data "$EVENT_DATA" \
        "$CALENDAR_SERVICE_URL" 2>&1 || echo "curl_error")

    if [ "$HTTP_CODE" != "200" ]; then
        log "ERROR: Failed to create event for '$TEST_NAME'. HTTP code: $HTTP_CODE"
        if [ -f "$RESPONSE_FILE" ]; then
            log "Response was:"
            cat "$RESPONSE_FILE"
            rm "$RESPONSE_FILE"
        fi
        return 1
    fi

    # 3. Extract Event ID
    CREATED_EVENT_ID=$(sed -n 's/.*"id":"\([^"]*\)".*/\1/p' "$RESPONSE_FILE")

    if [ -z "$CREATED_EVENT_ID" ]; then
        log "ERROR: Could not extract event ID for '$TEST_NAME'."
        log "Response was:"
        cat "$RESPONSE_FILE"
        rm "$RESPONSE_FILE"
        return 1
    fi

    log "Event created successfully with ID: $CREATED_EVENT_ID"
    rm "$RESPONSE_FILE"

    # 4. Wait for sender to process (only if we expect a notification)
    if [ "$SHOULD_FIND_LOG" = true ]; then
        log "Waiting ${WAIT_TIME}s for sender to potentially process the event..."
        sleep $WAIT_TIME
    else
        log "Short wait (${WAIT_TIME}s) to ensure processing cycle..."
        sleep $WAIT_TIME
    fi

    # 5. Check Sender Logs
    log "Checking sender logs for notification of event ID: $CREATED_EVENT_ID"
    SENDER_LOGS=$(docker logs --tail $LOG_LINES_TO_CHECK "$SENDER_CONTAINER_NAME" 2>/dev/null || true)
    
    MATCHING_LOG=$(echo "$SENDER_LOGS" | \
                   grep '"msg":"notification sent"' | \
                   grep "\"id\":\"$CREATED_EVENT_ID\"" || true)

    # 6. Evaluate Result
    local TEST_RESULT=0
    if [ "$SHOULD_FIND_LOG" = true ]; then
        if [ -n "$MATCHING_LOG" ]; then
            log "SUCCESS: Test '$TEST_NAME' PASSED. Found expected notification log entry."
            log "Log entry: $MATCHING_LOG"
            TEST_RESULT=0
        else
            log "FAIL: Test '$TEST_NAME' FAILED. Expected notification log entry NOT found."
            log "Dumping last $LOG_LINES_TO_CHECK lines of sender logs for debugging:"
            echo "$SENDER_LOGS"
            TEST_RESULT=1
        fi
    else # SHOULD_FIND_LOG is false
        if [ -n "$MATCHING_LOG" ]; then
            log "FAIL: Test '$TEST_NAME' FAILED. Unexpected notification log entry WAS found."
            log "Log entry: $MATCHING_LOG"
            log "Dumping last $LOG_LINES_TO_CHECK lines of sender logs for debugging:"
            echo "$SENDER_LOGS"
            TEST_RESULT=1
        else
            log "SUCCESS: Test '$TEST_NAME' PASSED. Correctly did NOT find notification log entry."
            TEST_RESULT=0
        fi
    fi

    # 7. Cleanup: Attempt to delete the event created in this test case
    # This happens regardless of test pass/fail to avoid cluttering the DB
    delete_event "$CREATED_EVENT_ID"

    return $TEST_RESULT
}

# --- Main Execution ---
log "Starting Notification Integration Tests..."

FAILURES=0
TOTAL_TESTS=0
FIRST_TEST_PASSED=false

# --- Test Case 1: Past event with notification ---
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test_case "Past event with notification" "$PAST_EVENT_DATETIME" "1800s" true; then
    FIRST_TEST_PASSED=true
else
    FAILURES=$((FAILURES + 1))
fi

# --- Test Case 2: Past event without notification ---
TOTAL_TESTS=$((TOTAL_TESTS + 1))
# Run this test regardless of the first test's outcome
if ! run_test_case "Past event without notification" "$PAST_EVENT_DATETIME" "0s" false; then
    FAILURES=$((FAILURES + 1))
fi

# --- Test Case 3: Future event ---
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if ! run_test_case "Future event" "$FUTURE_EVENT_DATETIME" "1800s" false; then
    FAILURES=$((FAILURES + 1))
fi

# --- Final Result ---
log ""
log "=== Test Suite Completed ==="
log "Total tests run: $TOTAL_TESTS"
log "Failures: $FAILURES"

if [ $FAILURES -eq 0 ]; then
    log "OVERALL RESULT: ALL TESTS PASSED"
    exit 0
else
    log "OVERALL RESULT: $FAILURES OUT OF $TOTAL_TESTS TESTS FAILED"
    exit 1
fi
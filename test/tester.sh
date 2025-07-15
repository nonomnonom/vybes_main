#!/bin/bash

# This script performs a full user flow: register, login, unlock wallet, and wallet operations.

# --- Configuration ---
BASE_URL="http://localhost:8080/api/v1"

# --- 1. Generate Random User Data ---
RANDOM_NUM=$RANDOM
USER_NAME="User ${RANDOM_NUM}"
USER_EMAIL="user${RANDOM_NUM}@example.com"
PASSWORD="password123"

echo "--- 1. Registering New User ---"
echo "Name: ${USER_NAME}"
echo "Email: ${USER_EMAIL}"

JSON_REGISTER_PAYLOAD=$(jq -n \
                  --arg name "$USER_NAME" \
                  --arg email "$USER_EMAIL" \
                  --arg password "$PASSWORD" \
                  '{name: $name, email: $email, password: $password}')

curl -s -X POST "${BASE_URL}/users/register" \
-H "Content-Type: application/json" \
-d "$JSON_REGISTER_PAYLOAD"
echo -e "\nRegistration complete.\n"


# --- 2. Login with the New User ---
echo "--- 2. Logging In ---"
JSON_LOGIN_PAYLOAD=$(jq -n \
                  --arg email "$USER_EMAIL" \
                  --arg password "$PASSWORD" \
                  '{email: $email, password: $password}')

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/users/login" \
-H "Content-Type: application/json" \
-d "$JSON_LOGIN_PAYLOAD")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" == "null" ]; then
    echo "Login failed. Could not get access token."
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo "Login successful. Tokens received."
echo -e "\n"


# --- 3. Unlock Wallet ---
echo "--- 3. Unlocking Wallet ---"
JSON_UNLOCK_PAYLOAD=$(jq -n \
                  --arg password "$PASSWORD" \
                  '{password: $password}')

UNLOCK_RESPONSE=$(curl -s -w "\nHTTP_STATUS_CODE:%{http_code}\n" -X POST "${BASE_URL}/wallet/unlock" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" \
-d "$JSON_UNLOCK_PAYLOAD")
echo "Unlock Response:"
echo "$UNLOCK_RESPONSE"
echo -e "\n"


# --- 4. Perform Authenticated Wallet Action (Personal Sign) ---
echo "--- 4. Performing Wallet Action (personal_sign) ---"
MESSAGE_TO_SIGN="Hello, Vybes! This is a test message."
JSON_SIGN_PAYLOAD=$(jq -n \
                  --arg message "$MESSAGE_TO_SIGN" \
                  '{message: $message}')

SIGN_RESPONSE=$(curl -s -w "\nHTTP_STATUS_CODE:%{http_code}\n" -X POST "${BASE_URL}/wallet/personal-sign" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" \
-d "$JSON_SIGN_PAYLOAD")
echo "Sign Response:"
echo "$SIGN_RESPONSE"
echo -e "\n"


# --- 5. Export Private Key (Authenticated) ---
echo "--- 5. Exporting Private Key (Authenticated) ---"
JSON_EXPORT_PAYLOAD=$(jq -n \
                  --arg password "$PASSWORD" \
                  '{password: $password}')

EXPORT_RESPONSE=$(curl -s -w "\nHTTP_STATUS_CODE:%{http_code}\n" -X POST "${BASE_URL}/wallet/export" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" \
-d "$JSON_EXPORT_PAYLOAD")
echo "Export Response:"
echo "$EXPORT_RESPONSE"
echo -e "\n"


# --- 6. Attempt to Export Private Key (Unauthenticated) ---
echo "--- 6. Attempting to Export Private Key (Unauthenticated) ---"
echo "This request is expected to fail with a 401 Unauthorized status."

UNAUTH_EXPORT_RESPONSE=$(curl -s -w "\nHTTP_STATUS_CODE:%{http_code}\n" -X POST "${BASE_URL}/wallet/export" \
-H "Content-Type: application/json" \
-d "$JSON_EXPORT_PAYLOAD")
echo "Unauthenticated Export Response:"
echo "$UNAUTH_EXPORT_RESPONSE"
echo -e "\n"

echo "--- Test Complete ---"
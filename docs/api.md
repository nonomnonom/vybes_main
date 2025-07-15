# Vybes API Documentation

This document provides detailed information about the Vybes API endpoints.

**Base URL**: `/api/v1`

---

## Authentication

All authenticated routes require a valid JWT to be sent in the `Authorization` header.

`Authorization: Bearer <your_jwt_token>`

---

## 1. User & Auth Endpoints

These endpoints handle user registration, login, profile management, and password recovery.

### `POST /users/register`
- **Description**: Registers a new user.
- **Request Body**:
  ```json
  {
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }
  ```
- **Response (201 Created)**: The newly created user object.

### `POST /users/login`
- **Description**: Logs in a user and returns a JWT.
- **Request Body**:
  ```json
  {
    "email": "test@example.com",
    "password": "password123"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "your_jwt_token"
  }
  ```

### `POST /users/request-otp`
- **Description**: Sends a One-Time Password (OTP) to the user's email for password reset.
- **Request Body**:
  ```json
  {
    "email": "test@example.com"
  }
  ```
- **Response (200 OK)**: `{"message": "OTP sent"}`

### `POST /users/reset-password`
- **Description**: Resets the user's password using a valid OTP.
- **Request Body**:
  ```json
  {
    "email": "test@example.com",
    "otp": "123456",
    "newPassword": "new_secure_password"
  }
  ```
- **Response (200 OK)**: `{"message": "Password reset successful"}`

### `GET /users/:username` (Auth Required)
- **Description**: Retrieves the profile of a specific user.
- **Response (200 OK)**: The user profile object.

### `PATCH /users/me` (Auth Required)
- **Description**: Updates the profile of the authenticated user.
- **Request Body**:
  ```json
  {
    "name": "New Name",
    "bio": "This is my new bio.",
    "profilePictureURL": "https://example.com/new_pfp.jpg"
  }
  ```
- **Response (200 OK)**: The updated user object.

---

## 2. Content & Post Endpoints

These endpoints handle the creation, deletion, and interaction with posts.

### `POST /posts` (Auth Required)
- **Description**: Creates a new post. This is a `multipart/form-data` request.
- **Form Data**:
  - `media`: The video file to upload.
  - `caption`: (Optional) The caption for the post.
  - `visibility`: (Optional) `public`, `friends`, or `private`. Defaults to `public`.
- **Response (201 Created)**: The newly created post object.

### `DELETE /posts/:postID` (Auth Required)
- **Description**: Deletes a post owned by the authenticated user.
- **Response (204 No Content)**

### `POST /posts/:postID/repost` (Auth Required)
- **Description**: Reposts an existing post.
- **Response (201 Created)**: The new repost object.

### `GET /reposts/by-user/:userID` (Auth Required)
- **Description**: Retrieves all reposts made by a specific user.
- **Response (200 OK)**: An array of post objects.

### `POST /posts/:postID/view`
- **Description**: Records a view for a post. This is a public endpoint.
- **Response (204 No Content)**

---

## 3. Interaction Endpoints (Comments, Likes, Bookmarks)

### `POST /posts/:postID/comments` (Auth Required)
- **Description**: Adds a comment to a post.
- **Request Body**:
  ```json
  {
    "text": "This is a great post!"
  }
  ```
- **Response (201 Created)**: The new comment object.

### `GET /posts/:postID/comments` (Auth Required)
- **Description**: Retrieves all comments for a specific post.
- **Response (200 OK)**: An array of comment objects.

### `POST /posts/:postID/like` (Auth Required)
- **Description**: Likes a post.
- **Response (204 No Content)**

### `DELETE /posts/:postID/like` (Auth Required)
- **Description**: Removes a like from a post.
- **Response (204 No Content)**

### `POST /posts/:postID/bookmark` (Auth Required)
- **Description**: Bookmarks a post.
- **Response (204 No Content)**

### `DELETE /posts/:postID/bookmark` (Auth Required)
- **Description**: Removes a bookmark from a post.
- **Response (204 No Content)**

### `GET /bookmarks` (Auth Required)
- **Description**: Retrieves all posts bookmarked by the authenticated user.
- **Response (200 OK)**: An array of post objects.

---

## 4. Social & Feed Endpoints

### `POST /users/:username/follow` (Auth Required)
- **Description**: Follows a user.
- **Response (204 No Content)**

### `DELETE /users/:username/follow` (Auth Required)
- **Description**: Unfollows a user.
- **Response (204 No Content)**

### `GET /feeds/for-you` (Auth Required)
- **Description**: Retrieves the "For You" feed for the authenticated user.
- **Response (200 OK)**: An array of post objects.

### `GET /feeds/friends` (Auth Required)
- **Description**: Retrieves the "Friends" feed (mutuals) for the authenticated user.
- **Response (200 OK)**: An array of post objects.

### `GET /suggestions/users` (Auth Required)
- **Description**: Gets a list of suggested users to follow.
- **Response (200 OK)**: An array of user objects.

---

## 5. Story Endpoints

### `POST /stories` (Auth Required)
- **Description**: Creates a new story. This is a `multipart/form-data` request.
- **Form Data**:
  - `media`: The image or video file for the story.
- **Response (201 Created)**: The new story object.

### `GET /stories/feed` (Auth Required)
- **Description**: Retrieves the story feed for the authenticated user.
- **Response (200 OK)**: An array of story objects grouped by user.

---

## 6. Notification Endpoints

### `GET /notifications` (Auth Required)
- **Description**: Retrieves notifications for the authenticated user.
- **Response (200 OK)**: An array of notification objects.

### `PATCH /notifications/read` (Auth Required)
- **Description**: Marks specified notifications as read.
- **Request Body**:
  ```json
  {
    "notificationIds": ["id1", "id2"]
  }
  ```
- **Response (204 No Content)**

---

## 7. Search Endpoints

### `GET /search/users` (Auth Required)
- **Description**: Searches for users by name or username.
- **Query Parameters**:
  - `q`: The search query.
- **Response (200 OK)**: An array of user objects.

---

## 8. Wallet Endpoints (Advanced)

These endpoints are for interacting with the user's self-hosted EVM wallet. They all require the user's account password for authorization in the request body.

### `POST /wallet/export` (Auth Required)
- **Description**: Exports the user's encrypted private key.
- **Request Body**: `{"password": "user_password"}`
- **Response (200 OK)**: `{"privateKey": "encrypted_private_key"}`

### `POST /wallet/personal-sign` (Auth Required)
- **Description**: Signs a message using `personal_sign`.
- **Request Body**: `{"password": "user_password", "message": "message_to_sign"}`
- **Response (200 OK)**: `{"signature": "0x..."}`

### `POST /wallet/sign-transaction` (Auth Required)
- **Description**: Signs an Ethereum transaction.
- **Request Body**: `{"password": "user_password", "transaction": {...}}`
- **Response (200 OK)**: `{"signedTx": "0x..."}`

*(Other wallet endpoints like `send-transaction`, `sign-typed-data`, and `secp256k1-sign` follow a similar pattern.)*
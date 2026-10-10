# AUTH-01: Customer Registration

## Endpoint

POST /api/auth/register
Content-Type: application/json

Creates an active customer account.

## Request example

```json
{
  "first_name": "Ahmed",
  "last_name": "Hussain",
  "email": "ahmed.registration@example.com",
  "phone_number": "07700900123",
  "password": "TestPass123!",
  "confirm_password": "TestPass123!"
}
```

## Validation

All six fields are required.

- First and last names: maximum 100 characters each.
- Email: valid address, maximum 254 bytes, must not already exist.
- Phone number: maximum 30 characters.
- Password: minimum 8 characters, maximum 72 bytes.
- Confirm password: must match the password.

The minimum password rule is provisional and needs team agreement.

Names and phone numbers are trimmed. Email is trimmed and converted
to lowercase. Passwords are hashed using bcrypt before saving.
Passwords and hashes are excluded from responses.

## Responses

| Status | Meaning |
|---|---|
| 201 | Account created successfully |
| 400 | Invalid JSON or unknown request field |
| 405 | Request method is not POST |
| 409 | Email is already taken |
| 422 | Validation failed; errors identify each invalid field |
| 500 | Database or password-processing failure |

Successful responses include first_name, last_name, email,
phone_number, role, and status.

## Running locally

Create a PostgreSQL database and set DATABASE_URL in the terminal.

Run the existing migration command once to create the tables:

```powershell
go run ./cmd/migrate
```

Start the API in the same terminal:

```powershell
go run .
```

Local address: http://127.0.0.1:8080

## Tests

```powershell
go test ./...
```

Registration tests use a temporary SQLite database without CGO.
They check successful registration, password hashing, safe responses,
duplicate emails, invalid fields, invalid JSON, and incorrect methods.

Manual PostgreSQL checks returned 201, 409, and 422 as expected.
The saved customer had customer role, active status, and a password
value different from the original password.

## Scope

This task covers customer registration only.
Login, JWT authentication, and logout are separate tasks.
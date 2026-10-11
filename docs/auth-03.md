# AUTH-03: Current user and logout

## Authentication contract

Requests use: Authorization: Bearer <token>

Login must issue an HS256 JWT signed using JWT_SECRET.
The sub claim must contain the user's database ID as a string.
The exp claim must contain the token's expiry timestamp.
JWT_SECRET must be a strong random secret shared by login and validation.
Secrets and database credentials must stay outside source code.

## GET /api/auth/me

Returns the current user's ID, first name, last name, email,
phone number, role and status. Password hashes are never returned.

Responses:
- 200: User profile returned.
- 401: Missing, invalid or expired token, or user no longer exists.
- 403: User account is not active.
- 500: Database lookup failed.

## POST /api/auth/logout

Requires a valid token belonging to an active user.
Returns 204 with no response body.

Logout does not invalidate the token on the server.
A retained copy remains valid until its expiry time.

The frontend must discard its token, clear the signed-in user state,
and redirect to login. It must clear local authentication state even
if the logout request fails, and prevent access to protected pages
without a token.

## Integration still required

Connect these endpoints to the team's login implementation.
Implement frontend token removal and protected-page routing.
Test using a real database user and a token issued by login.
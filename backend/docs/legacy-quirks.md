# Legacy PHP contract — characterization evidence

This document consolidates the live-characterization findings the Go
backend's code comments refer back to. It exists because in-code comments
previously pointed at "sdd/migrate-go-vue/..." paths in an external
project-memory store (Engram) that is not part of this repository — a dead
reference for any reader who does not have access to that store (gate #36
readability finding, PR3 corrective re-run).

Methodology (repeated for every resource ported so far): a scratch copy of
`FilasServer/` was pointed at the local Docker `db` service, patched only
for MySQL 8 case-sensitivity (`Admins`→`admins`, `Salt`→`salt`, etc. — the
legacy schema and legacy PHP code disagree on column/table case, which
MySQL 5.7 on the original host tolerated and MySQL 8 does not), run with
`php -S`, and exercised with `curl` for every documented operation. Raw DB
state was inspected via a `mysqli` one-liner after mutating calls to
confirm what was ACTUALLY persisted, not just what the HTTP response
claimed. Scratch files were never committed.

## 1. `SELECT *` password/salt leak (admins)

Legacy `admins.php`'s `getAdmins()` and `getUser()` both run `SELECT *
FROM Admins` and echo the raw row via `json_encode`. Live-confirmed: every
`GET /api/admins` and `GET /api/admins/:id` response included the admin's
password hash AND salt in plaintext JSON.

Go fix: `dto.AdminResponse` structurally has no `Password`/`Salt` field —
it cannot leak them by construction, regardless of what the repository
query selects internally.

## 2. `floatval($password)` update bug (admins)

Legacy `updateUser()` runs `floatval($data->password)` before storing it:
a non-numeric string like `"newpw"` becomes the number `0`, so `PUT
/api/admins/:id` silently replaced a working credential with an unusable
one. Live-confirmed via direct SQL inspection of the `password` column
after a PUT.

Go fix: `usecase.AdminService.Update` bcrypt-hashes the submitted password
before persisting — a real, working credential update.

**PR3 corrective addendum (blocker fix, gate #36):** the above fix
introduced a NEW vulnerability the floatval bug had accidentally been
blocking — an omitted/empty password reaching `auth.HashPassword("")`
produces a VALID, persistable bcrypt hash of the empty string. `Update` now
requires both `username` and `password` to be non-empty (mirroring
`Create`'s guard), returning `domain.ErrValidation` → 400 otherwise. This
is a deliberate divergence from legacy, which had no such guard at all
(see the doc comment on `usecase.AdminService.Update`).

## 3. bcrypt + transparent migrate-on-login (admins)

Legacy hashes passwords as `sha256(salt + rawPassword)` with a random
salt per admin (`admins.php::checkUserForLogin`/`createUser`). Go upgrades
to bcrypt with a transparent migration: `usecase.AuthService.Login` tries
bcrypt first; if the stored hash isn't bcrypt-shaped, it falls back to
reproducing the legacy scheme (`internal/auth/password.go`,
`VerifyLegacySHA256`, constant-time compare). A successful legacy
verification re-hashes to bcrypt and persists it — the admin never notices,
no forced reset.

Live-verified end-to-end: first login against the seeded synthetic admin
(`FilasAdmin` / `filas-local-dev-2026`, seeded with the legacy scheme)
migrated the stored hash to `$2a$...`; a second login with the same
credentials left the hash unchanged (idempotent, not re-hashed every
login).

**PR3 corrective addenda (gate #36 risk findings):**
- **Timing oracle:** the user-not-found path used to return near-instantly
  while a wrong-password (known-username) attempt burned a real bcrypt
  compare (~50-100ms) — a measurable, attacker-observable timing
  difference that leaks which usernames exist. `Login` now performs a
  dummy bcrypt comparison against a fixed, precomputed hash on the
  not-found path (see `dummyBcryptHash` in `usecase/auth.go`) so both
  paths cost about the same.
- **TOCTOU:** the migrate-on-login rehash persisted with a plain
  `UPDATE ... WHERE ID = ?`, i.e. unconditional last-write-wins — a
  concurrent password rotation could be silently reverted to a bcrypt hash
  of the OLD password. `AdminRepository.UpdatePassword` now takes the
  exact old hash that was read and verified and conditions the UPDATE on
  it (`... WHERE ID = ? AND password = ?`); if the row changed since the
  read, the statement matches 0 rows and is a no-op by design.
- **Silent failure:** a hashing or persistence failure during
  migrate-on-login used to be discarded (`_ = s.repo.UpdatePassword(...)`).
  Both failure modes are now logged (`log.Printf`, `usecase/auth.go`) so a
  persistently-failing migration is observable server-side instead of
  vanishing.

## 4. Admin ID assignment (admins)

Legacy `createUser()` required the CALLER to supply `ID` in the request
body and inserted it explicitly. The seed schema patches `admins` with
`PRIMARY KEY (ID) AUTO_INCREMENT`; `AdminRepository.Create` no longer
accepts a caller-supplied ID and reads `LastInsertId()` instead. Live
end-to-end verified: a create call with no `ID` in the body was assigned
one by the database.

## 5. Unified 401 (admins auth gating)

Legacy `admins.php` distinguishes a MISSING `Authorization` header (400,
empty body) from an INVALID token (401, `{"message":"Token is invalid"}`)
on every gated route. Go's `middleware.RequireAuth` (and the inline
equivalent used by `POST /api/admins`'s non-login branch, see §7 below)
unify BOTH cases to 401: missing header →
`{"message":"missing authorization token"}`, invalid/expired token →
`{"message":"invalid or expired token"}`. This is a deliberate, documented
simplification (not a legacy-parity requirement any decision covers) —
reusing the ONE auth contract established for every other gated route
rather than forking a second status-code scheme just for admins.

## 6. `admins` GET routing without a trailing slash

Legacy's manual `explode('/')` URL parsing has a bug: `GET /api/admins`
(no trailing slash, no ID) evaluates `$parts[4] !== ""` as true even
though `$parts[4]` is undefined, misrouting into `getUser(0)` → 404, even
though the caller wanted the list. This is NOT reproduced in Go. It is an
artifact of PHP's manual string-splitting router, not exercised by the
real frontend client (`API.mjs` always calls with a trailing slash, the
case legacy resolves correctly to the list). Go's `:id` path-param routing
has no equivalent failure mode — there is nothing to intentionally NOT
reproduce, because the bug class does not exist here. See `router.go` for
the route registration.

## 7. Shared bearer-auth logic

`middleware.RequireAuth` (route-group middleware, used by every gated
GET/PUT/DELETE route) and `AdminHandler.requireAuthInline` (used only by
`POST /api/admins`'s non-login branch, since the same route/method also
serves the public login call and cannot be gated at the route-group level)
both need to: read the `Authorization` header, tolerate an optional
`Bearer ` prefix (the legacy client sends the raw token, no prefix), parse
the JWT, and emit one of the same two 401 bodies on failure. This logic
used to be hand-duplicated at both call sites (gate #36 readability
finding — a real risk, since the two literal 401 message strings could
silently drift). It now lives in exactly one place,
`middleware.AuthenticateRequest`, which both call sites invoke.

## 8. PR2 public-GET resources — per-resource response quirks

Each of the five public GET resources (`products`, `news`, `gallery`,
`family`, `organizations`) was characterized independently; they do NOT
all share the same shape:

| Resource | Empty list | By-ID hit shape | Notes |
|---|---|---|---|
| products | 200 `[]` | bare object | dedicated `getProducts()`/`getProduct()` functions, own code path |
| news | 404 | **array-wrapped** (`[{...}]`) | by-ID reuses the list row-collector loop |
| gallery | 404 | array-wrapped | same shared row-collector pattern as news |
| family | 404 | array-wrapped | same shared row-collector pattern as news |
| organizations | 404 | array-wrapped | same shared row-collector pattern as news |
| admins (PR3) | 200 `[]` | bare object | own dedicated functions, same pattern as products |

Additionally, nullable text columns (`Product.Image`/`Description`,
`News.Body`/`Image`, `Family.Image`, `Organization.Description`) are
mapped through `sql.NullString` in every repository — scanning a NULL
directly into a plain Go `string` errors, which used to 500 the whole
endpoint for any row with a NULL image (characterized live by inserting a
NULL-image row and confirming a 500, then confirming 200 after the fix).
`AdminRepository` uses the same defensive `sql.NullString` mapping for
consistency even though a real admin row never actually has a NULL
username/password/salt.

## 9. `intval()` mixed-string ID parsing — READS only (PR2, all resources)

Scope: this entry covers **GET** requests only. See §10 below for the
distinct, separately-characterized behavior of non-numeric IDs on
**write** (PUT/DELETE) requests, which does NOT resolve to "not found".

PHP's `intval("5abc")` parses a leading numeric prefix out of a mixed
string and returns `5`. `parseLegacyID` (`internal/handler/rest/id.go`)
does not replicate that: any path segment that isn't fully numeric returns
`0` (an ID that never exists for GET, resolving to the same "not found"
outcome intval's stricter cases produce, since `Get(ctx, 0)` legitimately
finds no row and the handler maps `domain.ErrNotFound` -> 404). Deliberate,
documented simplification — no real client (the vanilla-JS frontend) ever
sends a mixed numeric/alpha ID segment.

## 10. Write ops (PUT/DELETE) on a missing or non-numeric ID return 200, NOT 404 (PR4 corrective)

This is EXACT legacy parity, not a bug to fix — it was previously
undocumented and untested (gate finding, PR4 corrective batch).

Legacy `updateProduct`/`deleteProduct` (`FilasServer/products.php:120-140`
and `:142-154`) check only `if ($conn->query($sql) === TRUE)`. PHP's
`mysqli::query()` returns `TRUE` for ANY successful DML statement,
including an `UPDATE`/`DELETE` that matches **zero rows** — mysqli only
returns `FALSE` on a genuine SQL/connection error, never on "no rows
matched". Consequently `PUT`/`DELETE /api/products/{missing-or-non-numeric-id}`
returns `200 {"message":"Product updated/deleted successfully"}` with no
actual mutation, exactly like a real ID would.

Live-characterized (PR4 corrective, scratch PHP against the seeded
Docker `db`, methodology as above): `PUT /api/products/999999` and
`DELETE /api/products/999999` (a guaranteed-nonexistent ID) both returned
`200` with the standard success message — confirmed live, not just by
code inspection.

Go reproduces this exactly: `ProductRepository.Update`/`.Delete` (and the
equivalent methods on every other resource's repository) execute an
`UPDATE`/`DELETE ... WHERE ID = ?` and only return an error on a genuine
`database/sql` exec error — a zero-rows-affected result is not an error
in Go's `database/sql` any more than it is in mysqli, so no special-casing
was needed to match. `parseLegacyID` returning `0` for a non-numeric path
segment (see §9) therefore also resolves to a 200 no-op on write routes,
not a 404 — the two write handlers (`ProductHandler.Update`/`.Delete`,
and their siblings) never check `domain.ErrNotFound` at all, unlike
`Get`, which explicitly does. Locked by
`TestProductHandler_Update_TreatsNonNumericIDAsNoOpSuccess` /
`_Delete_...` (handler layer) and
`TestProductRepository_Update_SucceedsWithZeroRowsAffected` /
`TestFamilyRepository_Update_SucceedsWithZeroRowsAffected` (repository
layer, `sqlmock.NewResult(0, 0)`).

## 11. Description excluded from `updateProduct`'s SQL (products)

Live-characterized (PR4 corrective; also visible directly in
`FilasServer/products.php:120-140`): `updateProduct()`'s `UPDATE` statement
(lines 129-131) is `UPDATE products SET Name = '$name', Price = $price,
Stock = $stock, Image = '$image' WHERE ID = $id` — it never references the
`Description` column at all, unlike `createProduct()`'s `INSERT`, which
does. A `PUT /api/products/:id` can therefore never change a product's
Description, regardless of what the request body contains; only the value
set at `POST` time (`createProduct`) persists.

Confirmed live: created a product with `Description:"original-description"`,
then `PUT` the same product with a different `Description` in the body —
the raw DB row afterward still held `"original-description"` untouched,
while Name/Price/Stock/Image all reflected the PUT.

Go reproduces this: `domain.ProductRepository.Update` and
`mysql.ProductRepository.Update`/`productUpdate` (the SQL constant)
deliberately omit `Description` from the `UPDATE` statement — see their
doc comments, which now cite this entry.

## 12. Orders' `GROUP_CONCAT` JSON shape (not yet ported)

Not yet characterized in code — flagged in the design doc (§9) as a risk
for the orders PR: the legacy response is built with `GROUP_CONCAT` plus
manual JSON string concatenation, which is fragile with quotes/commas in
product names. The Go DTO must reproduce the SHAPE, not the bug's
fragility. Tracked for the orders work unit, not covered by this document.

## 13. News write auth bug — FIXED, not reproduced (PR5, task 4.1)

Live-characterized (also visible directly in `FilasServer/news.php:30-33`,
repeated identically at `:61-64` for PUT and `:94-97` for DELETE). The
auth gate on every news write is:

```php
$headers = getallheaders();
if(isset($headers['Authorization'])) {
    $token = trim(str_replace('Bearer ', '', $headers['Authorization']));
    TokenValidationResponse($token);
}
```

Two independent bugs compound here, both live-confirmed against a scratch
copy of `FilasServer/` pointed at the local Docker `db` (per this
document's Methodology note) with a 3-hour-old fixed news seed as the
baseline:

1. **No header -> no check at all.** The `isset($headers['Authorization'])`
   guard means a request with NO `Authorization` header never calls
   `TokenValidationResponse` — the write proceeds unconditionally. Confirmed
   live: `POST /api/news` with no `Authorization` header at all returned
   `201 {"message":"News added successfully"}`, and the row was verified
   present in the database (`SELECT * FROM news` showed the new row with
   the test `Title`).
2. **Even when a header IS present, the validation result is discarded.**
   `TokenValidationResponse($token)`'s return value is never checked, and
   the function itself does not `exit`/`die` on failure — it only `echo`s
   `{"message":"Token is invalid"}` and returns `false`, which the caller
   ignores. Confirmed live: `POST /api/news` with
   `Authorization: Bearer not-a-real-token` (garbage, unparseable) returned
   HTTP 201 with a body of **two concatenated JSON objects**,
   `{"message":"Token is invalid"}{"message":"News added successfully"}`
   — direct evidence the code kept executing and the row was persisted
   (verified in the database) despite the token being invalid.

Both characterization rows were deleted after confirmation; the scratch
`FilasServer/` copy and its patched `index.php`/`news.php` were never
committed and were deleted after this session (`docker compose down` was
NOT required, since the scratch PHP connected to the already-running `db`
service standalone; only the scratch directory and the temporary `php -S`
process were torn down).

**Per user decision (obs #25) and spec requirement ("News Write
Authentication (MODIFIED, CHANGE)"), this is NOT reproduced.** The Go
backend applies `middleware.RequireAuth` unconditionally to all three news
write routes (`router.go`), exactly like every other gated resource: a
missing OR invalid/expired token always returns 401 before the usecase or
repository is ever invoked — no mutation, no double-JSON body, no silent
fall-through. `NewsHandler.Create`/`Update` also depart from the generic
`middleware.ErrorHandler` validation-message convention (see
`usecase.validateNewsItem` and `NewsHandler`'s `newsMissingFieldMessage`
doc comments): the 400 body for a missing `Title`/`Body`/`Image` is the
exact legacy string `"Missing Title, Body, or Image parameter"`, verbatim,
unlike products/gallery/family's informally-worded Go validation messages
(see §10/§11), because reproducing that particular string was a stated
parity requirement for this endpoint while the AUTH behavior around it was
explicitly NOT.

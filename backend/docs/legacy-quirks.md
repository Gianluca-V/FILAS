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

## 12. Orders' `GROUP_CONCAT` JSON shape — see §15

Superseded: this entry originally flagged the orders `GROUP_CONCAT` shape
as "not yet characterized" (design doc §9 risk, tracked for a future PR).
It has since been live-characterized and fully documented in §15.2, along
with the auth model (§15.1) and the trigger-case bug (§15.3). This section
number is kept as a redirect stub because §14 cites "§12" by number; follow
that reference to §15 for the actual content.

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

## 14. Orders — 5 legacy triggers ported to `usecase/order.go` (PR6, tasks 5.1/5.4)

Order business logic in legacy was split across two places that had to be
read together to understand actual behavior: five MySQL triggers on
`orderproduct`/`orders` (`01-schema.sql`), and three PHP entry points in
`FilasServer/orders.php` (`createOrder`, `updateOrder`, `patchOrder`).
`usecase.OrderService` (domain: `internal/domain/order.go`) ports both into
one place, per design decision #2 (ADR-3). The triggers themselves stay in
the seed, unexecuted-by-Go, until PR7's gate drops them once the order
repository/handler land and characterization tests are green.

**Trigger -> usecase mapping:**

| Legacy trigger | Ported to |
|---|---|
| `before_orderProduct_insert` / `before_orderProduct_update` (`orderPrice = products.price * NEW.productQuantity`) | `OrderService.Create`/`Update`: `price := product.Price * float64(item.Quantity)` |
| `after_orderProduct_insert` / `after_orderProduct_update` (`orders.total = SUM(orderPrice) WHERE orderID = ...`) | `OrderService.Create`/`Update`: `total` accumulated as each line's price is computed |
| `before_update_orders` (`IF NEW.state='finished' AND OLD.state!='finished' THEN NEW.finishDate = CURRENT_TIMESTAMP`) | `OrderService.PatchState`: stamps `finishDate = time.Now()` only when `target == OrderStateFinished`, left `nil` otherwise |

**PHP entry point -> usecase mapping**, and the three DELIBERATE
divergences (confirmed against the design doc's §2.7/§9 and the spec's
"Public Checkout with Tightened Validation" / "Stock Non-Negative
Invariant" requirements before implementing):

1. **`createOrder` (orders.php:168) -> `OrderService.Create`.** Requires
   `orderProducts`/`name`/`phone` (else `ErrValidation`, mirrors the 400
   "Parameters missing" message). Per line: computes orderPrice and
   aggregates per-product demand, then hands the repository the order +
   its lines + the computed stock deltas in ONE atomic call (see the
   "Atomicity contract" subsection below) — ported from the inline
   `$newStock = $currentStock - $quantity` (orders.php:211), but now
   guarded:
   - **Quantity > 0, not just >= 0.** Legacy only rejected `quantity < 0`
     (orders.php:192), so `quantity == 0` silently passed despite the
     error message reading "Quantity can not be less than 1" — a bug
     against its own stated intent. The Go usecase rejects `quantity <= 0`.
   - **`ErrInsufficientStock`.** Legacy applied the stock subtraction
     unconditionally, with no floor check — directly responsible for the
     seed's `orderproduct` row 63 (`productQuantity=-100000`,
     `orderPrice=-85000000`) and `orders` row 31 (`total=-85000000`,
     `01-schema.sql`). The Go usecase rejects the WHOLE order if any
     distinct product's AGGREGATED demand across all its lines exceeds
     current stock, before calling the repository at all.
2. **`updateOrder` (orders.php:232) -> `OrderService.Update`.** Recomputes
   orderPrice/total for the replacement line items (same computation as
   Create) and replaces them via `OrderRepository.ReplaceProducts`.
   Quantity <= 0 is rejected for consistency with Create's invariant (not
   a legacy behavior — legacy's `updateOrder` had zero quantity
   validation). Deliberately does **NOT** touch product stock: legacy
   never adjusted stock on PUT either (a pre-existing gap the design doc
   §2.7 calls "the PUT stock-drift bug"), and properly reconciling it
   would require reading the order's OLD line items atomically with the
   new write — that needs the transactional repository landing in PR7,
   not the usecase alone operating on non-transactional reads. Locked by
   `TestOrderService_Update_DoesNotTouchStock` so a future change to this
   scope is a deliberate, reviewed decision, not a silent behavior shift.
3. **`patchOrder` (orders.php:261) -> `OrderService.PatchState`.**
   Preserved as-is (not a divergence): target state must be `finished` or
   `canceled` (else `ErrValidation`, 400 "invalid state", via
   `domain.ValidOrderTransitionTarget`); the order must currently be
   `pending` (else `ErrConflict`, 409 "order not pending"). On
   `pending -> canceled`, restores each DISTINCT product's aggregated
   line-item quantity back to stock — ported from the inline restore loop
   (orders.php:293-325), symmetric with Create's guarded decrement, and
   applied atomically with the state flip (see below).

### Atomicity contract — corrective (PR6 gate re-run)

The first PR6 pass orchestrated persistence as SEPARATE calls from the
usecase: `orders.Create(order)` followed by a loop of
`products.Update(...)` for Create; `orders.UpdateState(...)` followed by a
restore loop for PatchState's cancel path. A 3-lens gate review (risk +
resilience + reliability) failed this shape for three converged reasons —
an order could persist with only SOME of its stock decremented if the loop
failed partway (orphan order), a cancel could flip state without fully
restoring stock (partial restore), and a read-then-write state check
(`if order.State != pending`) was a check-then-act race two concurrent
PATCH calls could both pass — plus a **separate, pure-logic bug**: two
line items referencing the SAME product each validated against the SAME
un-mutated `products.Get()` snapshot independently, so their COMBINED
demand was never checked (e.g. two `{ProductID:1, Quantity:3}` lines vs
`Stock:5` each individually "fit" but 3+3=6 > 5 — a lost update that would
have driven stock negative exactly like the divergence #1 fix was meant to
prevent).

**Fix, in two parts:**

1. **Aggregate-before-check.** `usecase.aggregateStockDeltas` sums each
   line's quantity per DISTINCT `ProductID` before any stock check or
   delta is built — both `Create` (decrement) and `PatchState`'s cancel
   path (restore) call it. See
   `TestOrderService_Create_RejectsInsufficientStockForCombinedDuplicateDemand`
   and `TestOrderService_PatchState_AggregatesDuplicateProductLinesOnCancel`.
2. **Reshape `domain.OrderRepository` around atomic composite operations.**
   The usecase now computes a PLAN (prices, total, aggregated
   `[]domain.StockAdjustment` deltas, finishDate) and hands it to the
   repository as ONE call per mutation:
   - `Create(ctx, order, deltas)` — MUST atomically insert the order, its
     lines, AND apply every delta in a single transaction (PR7). No
     separate `products.Update()` loop from the usecase.
   - `TransitionState(ctx, orderID, target, finishDate, deltas)` — MUST
     atomically perform a CONDITIONAL update (`UPDATE orders SET ... WHERE
     ID=? AND state='pending'`) and, only if that succeeds, apply
     finishDate/deltas in the SAME transaction. Returns `ok=false, err=nil`
     when the conditional update affected zero rows — the CAS failure the
     usecase maps to `domain.ErrConflict`, closing the TOCTOU gap: the
     usecase no longer authorizes the transition itself from a stale
     `Get()`, it only asks the repository to attempt it atomically and
     trusts the boolean result. See
     `TestOrderService_PatchState_ReturnsConflictWhenTransitionNotOK`.

Both methods' doc comments on `domain.OrderRepository` spell out this
atomicity/CAS contract explicitly so PR7 cannot land a non-atomic
implementation that still happens to pass the usecase's unit tests (which
only exercise the CONTRACT via fakes, not real transactional behavior).

**Not yet built (PR7 scope, explicitly out of bounds for this port):**
`repository/mysql/order.go` (the transactional repo implementing
`domain.OrderRepository`'s atomic `Create`/`TransitionState` contracts —
see above), `handler/rest/order.go` + DTOs (including the `GROUP_CONCAT`
JSON shape from §12), router wiring, and dropping the 5 triggers from
`01-schema.sql`. `usecase.OrderService` and `domain.OrderRepository` are
verified with table-driven unit tests against fake repositories only —
there is no live/Docker verification for this PR, since there is no
DB-backed order path yet to verify against; the fakes prove the usecase
calls the atomic contract correctly, not that PR7's real implementation
IS atomic (that is PR7's own test responsibility).

## 15. Orders repo/handler + GROUP_CONCAT shape + trigger-case bug (PR7, tasks 5.2/5.3/6.1)

Live-characterized (same Methodology as this document's header note): a
scratch copy of `FilasServer/` was patched to connect to the local Docker
`db` on `127.0.0.1:${DB_PORT}`, with table-name case fixed for MySQL 8/
Linux (`orderProduct` -> `orderproduct` in `orders.php`'s own SQL, `Products`
-> `products`) and exercised with `curl` against every documented
operation, including mutating calls verified against the raw DB state
afterward. Scratch files were never committed.

### 15.1 Orders auth model — PUBLIC create, gated everything else

Confirmed directly in `FilasServer/orders.php`'s routing switch: `POST`
(`createOrder`) has **NO** auth check at all — no `getallheaders()` call,
no `TokenValidationResponse`, nothing. `GET`/`PUT`/`PATCH` all gate on
`isset($headers['Authorization'])` (400 if absent) then
`TokenValidationResponse` (401 if invalid) — same two-tier scheme as
admins (§5), which Go unifies to a single 401 for both cases, same
precedent. This is a genuine, deliberate legacy design (not a bug like
news' auth hole, §13): `POST /api/orders` is the customer-facing checkout
endpoint and was never meant to require a login. Go reproduces this
exactly: `router.go` wires `orders.Create` with NO middleware, and
`orders.List`/`.Get`/`.Update`/`.PatchState` behind
`middleware.RequireAuth`.

### 15.2 `getOrders`/`getOrder` response shape — GROUP_CONCAT decoded live

Live-captured exact shape (order 27, `GET /api/orders/27` with a valid
token):

```json
[{"orderID":"27","orderTotal":"13100","orderState":"pending","orderStartDate":"2023-11-20 18:30:02","orderFinishDate":null,"orderName":"Turco agustin","orderPhone":"32142432546","products":[{"productName":"Mermelada de tomate","productPrice":600,"productQuantity":5}]}]
```

Key findings, all locked by `dto.OrderResponse`/`OrderProductResponse` and
the `TestContract_Orders*` fixtures in `backend/testdata/contract/`:

- `orderID`/`orderTotal`/`orderState`/`orderName`/`orderPhone` are JSON
  **strings** (same mysqli-string convention as every other resource,
  §8), even though `orderTotal` is numeric.
- `orderStartDate`/`orderFinishDate` are `"YYYY-MM-DD HH:MM:SS"` strings
  (MySQL's DATETIME-to-string form, no timezone, no "T" separator);
  `orderFinishDate` is JSON `null` when the column is NULL.
- A by-ID `GET` wraps the single result in a one-element array, same
  quirk as news/gallery/family/organizations (§8) — legacy's `getOrder`
  collects into `$order[]` even though at most one order can match.
- **`products[]` entries are genuine JSON NUMBERS for `productPrice`/
  `productQuantity`** (unquoted) — the ONLY place in this entire API where
  a numeric value isn't a JSON string. This is because they come from
  `json_decode(stripslashes($row['products']), true)` on a hand-built JSON
  STRING (the `GROUP_CONCAT` result), not from a raw mysqli column value —
  `json_decode` parses unquoted numeric literals as native PHP
  int/float, and `json_encode` re-emits them as JSON numbers.
- **CRITICAL, easy-to-get-wrong finding: `productPrice` is `p.price` — the
  product's CURRENT price via a live `JOIN` — NOT `op.orderPrice` (the
  frozen `price*quantity` actually billed at order time).** Live-confirmed
  by `PUT`-replacing order 27's line to `{productID:3, quantity:5}`
  (product 3's price is 600, so `orderPrice`=3000) and then `GET`-ing it
  back: the response showed `"productPrice":600` (product 3's unit price),
  never `3000` (the line total). A product whose price changes after an
  order is placed will show the NEW price on every subsequent `GET` of
  that historical order — this is preserved exactly, not fixed. See
  `domain.OrderProduct.ProductName`/`.ProductPrice` doc comments and
  `mysql.OrderRepository`'s `orderJoinSelect` (`p.Price AS productPrice`,
  NOT `op.orderPrice`).
- **`GET /api/orders` empty-list is 200 `[]`, NOT 404** — unlike news/
  gallery/family/organizations. Legacy's `getOrders()` checks
  `$result === false`, and a zero-row `SELECT` (via `$conn->query()`)
  returns a valid `mysqli_result` object, never `false` — same reasoning
  as products'/admins' 200-empty-list behavior (§8). Live-confirmed:
  `TRUNCATE`-ing `orders`/`orderproduct` and calling `GET /api/orders/`
  with a valid token returned `200 []`.
- **The underlying query is an INNER JOIN, not a LEFT JOIN**: an order
  with ZERO `orderproduct` rows is invisible to BOTH `getOrders` and
  `getOrder` — `getOrder`'s `$result->num_rows > 0` check on the JOINed
  result set is false, so it 404s with `{"message":"Order not found"}`
  even though the order row genuinely exists in `orders`. Every order
  created through `usecase.OrderService.Create`/`.Update` always has >=1
  line item, so this only matters for pre-existing garbage data. Go's
  `mysql.OrderRepository.List`/`.Get` reproduce this via the same INNER
  JOIN (`orderJoinSelect`), not a defensive LEFT JOIN.
- **Products array order is UNDEFINED in legacy, not part of the parity
  contract.** `GROUP_CONCAT` has no `ORDER BY` clause inside it
  (`orders.php:84-88`), so MySQL is free to return matching rows in
  storage/index order — live-confirmed: order 10's 3 line items came back
  in the sequence naranja-inglesa/pera/tomate, NOT `orderproduct.ID`
  ascending (16,17,18 -> pera/naranja/tomate) as a naive reading of the
  INSERT order would suggest. Go's `mysql.OrderRepository` uses a STABLE,
  deterministic `ORDER BY o.ID, op.ID` instead — a reasonable, documented
  choice given legacy's own order is unreliable, not a byte-parity target.
- **Deliberately NOT reproduced: the GROUP_CONCAT/manual-JSON fragility
  bug.** Legacy's `products` field is built via string `CONCAT(...)` —
  inserting a product with `Name = 'Dulce "Especial", con comas'` and
  ordering it, then `GET`-ing that order back, returned
  `"products":null` (live-confirmed): the embedded `"` breaks the
  hand-built JSON string, `json_decode` silently fails and returns `null`,
  and legacy blindly assigns that `null` to `$row['products']`. Go's
  `dto.OrderResponse` builds the array structurally via `encoding/json`,
  which properly escapes quotes/commas — `products` is ALWAYS a valid
  array regardless of product name content. Locked by
  `TestContract_Orders_ProductNameWithQuoteAndCommaDoesNotBreakTheResponse`
  / `orders_get_quote_robustness.json`. Per the design's own precedent
  (§9 "Architectural risks"): reproduce the SHAPE, not the bug's
  fragility.

### 15.3 Pre-existing trigger case-sensitivity bug, discovered and fixed (blocking, not scope creep)

While reaching the "triggers still present" half of the PR7 gate (task
6.1 step 1), every `INSERT INTO orderproduct` — whether issued by the
scratch PHP OR by a raw `mysql` client — failed with
`ERROR 1146: Table 'filas.orderProduct' doesn't exist`, even though the
query text itself said `orderproduct` (lowercase, matching the actual
table). Root cause: the `after_orderProduct_insert`/
`after_orderProduct_update` trigger BODIES (not the PHP, not the INSERT
statement) reference `FROM orderProduct` (mixed case) — copied verbatim
from the source `filas.sql` dump. MySQL 5.7/Windows (the original legacy
host) tolerates this because `lower_case_table_names` there is
case-insensitive; `mysql:8` on Linux (this project's Docker image) runs
with `lower_case_table_names=0` (case SENSITIVE), so the trigger's
internal `SELECT ... FROM orderProduct` genuinely cannot find the table
and the ENTIRE INSERT statement aborts.

This was a **latent, previously-undiscovered bug**: PR1-PR6 never
exercised a real INSERT into `orderproduct` against the dockerized DB (PR6
was fakes-only), so nothing had ever fired these triggers before now. It
made the "characterize WITH triggers present" half of the PR7 gate
literally impossible to execute (every order-creation attempt 500s), so
it was fixed as a narrowly-scoped, pre-existing-bug correction — `01-schema.sql`'s
two `AFTER` trigger bodies now read `FROM orderproduct` (lowercase) — done
BEFORE and SEPARATELY from the trigger-DROP itself (task 6.1 step 2-3).
Confirmed live: identical `INSERT INTO orderproduct (...)` 500s on the
unpatched schema, succeeds on the patched one. See `01-schema.sql`'s
header note 4 for the schema-file-level documentation of this fix.

### 15.4 Trigger-drop gate evidence

The gate proves the Go `usecase/order.go` + `repository/mysql/order.go`
path produces IDENTICAL persisted behavior whether the 5 legacy triggers
are present or absent — i.e. Go now fully owns orderPrice/total/finishDate/
stock, not the database. It was run live against the dockerized stack
(`docker compose up`, fresh seed) by driving the Go API and reading the raw
DB with `mysql` after each step.

The SAME order was created through `POST /api/orders` in both phases:
`{orderProducts:[{productID:1,quantity:2},{productID:23,quantity:3}]}`
(product 1 price 600, product 23 price 850).

| Observable (raw DB) | Phase A: 5 triggers PRESENT (case-fixed) | Phase B: 0 triggers (dropped) |
|---|---|---|
| `orderproduct.orderPrice` (product 1) | 1200 (= 600 × 2) | 1200 |
| `orderproduct.orderPrice` (product 23) | 2550 (= 850 × 3) | 2550 |
| `orders.total` | 3750 (= SUM) | 3750 |
| `products.stock` delta (p1 / p23) | −2 / −3 | −2 / −3 |
| `PATCH ->finished` stamps `finishDate` | yes, non-null | yes, non-null |
| `PATCH ->canceled` restores stock, leaves `finishDate` NULL | yes | yes |
| `SELECT COUNT(*) information_schema.triggers` | 5 | 0 |

Phase B is the decisive half: with ZERO triggers in the database, the Go
API still wrote `orderPrice`/`total` correctly and stamped `finishDate` on
the pending→finished transition (a different wall-clock value than Phase A,
proving Go — not `before_update_orders` — set it). A full `GET /api/orders/
:id` of the Go-created finished order was then captured from BOTH the Go
backend and a scratch copy of legacy `orders.php` pointed at the same
Docker DB, and the two responses were **byte-identical** (`diff` empty,
matching sha256) — closing the loop create→persist→read across both stacks.
The four `orderproduct` triggers and `before_update_orders` were removed
from `01-schema.sql` only after this evidence was green (see that file's
header note 1 and the inline drop notes).

### 15.5 Stock decrement is race-safe (floor-guarded CAS)

`mysql.OrderRepository.Create` decrements stock with a FLOOR-GUARDED
conditional UPDATE — `UPDATE products SET Stock = Stock + ? WHERE ID = ?
AND Stock + ? >= 0` (`stockDecrement`) — and treats `RowsAffected == 0` as
`domain.ErrInsufficientStock`, rolling the whole order back. This is the
authoritative, race-safe availability check; the usecase's pre-transaction
`Stock < demand` check is a fast/clear-failure optimization only (the two
together are defense-in-depth). It mirrors `TransitionState`'s CAS idiom.
Without this guard the usecase check is a TOCTOU: two concurrent
`POST /api/orders` for the same product both read the same stale stock and
both decrement, overselling into negative stock (this was the PR7 review
gate's converged blocker — the same integrity family as PR6's
duplicate-line lost update, but across concurrent distinct orders).

Live-verified against the dockerized stack: with `products.Stock = 3` and
12 concurrent `POST /api/orders` (qty 1 each), exactly **3** succeeded,
final `Stock` landed at **0 (never negative)**, and the surplus requests
were rejected — including at least one via the repository CAS (`RowsAffected
== 0 → 409`) after its stale pre-check passed, proving the race is closed at
the persistence boundary, not just the read.

**Known limitation (tracked follow-up, not a blocker for this local-only
project):** under that same high concurrent contention on a single product
row, some transactions are chosen as InnoDB deadlock victims (Error 1213)
and surface as `500`. This is a PRE-EXISTING lock pattern, not introduced by
the CAS: the `orderproduct` insert takes a shared lock on the referenced
`products` row (FK `orderproduct_ibfk_2`), and the subsequent stock UPDATE
needs an exclusive lock on that same row — an S→X upgrade that deadlocks
when several checkouts of the same product interleave. It is
integrity-preserving (the victim rolls back atomically — no oversell, no
partial write, confirmed by `Stock = 0` above) and retriable by design
(Error 1213 explicitly says "try restarting transaction"). A production
system would wrap `Create` in a bounded deadlock-retry; that is deliberately
deferred here as out of scope for a single-operator local demo.

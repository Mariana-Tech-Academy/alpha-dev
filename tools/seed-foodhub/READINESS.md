# Original five-restaurant catalogue and historical readiness notes

The owner has confirmed the existing local FoodHub demo is working. This file
preserves the original five-restaurant menu reference and preparation findings;
it is not a current database inventory or an instruction to repeat completed setup.
The current catalogue also includes Tokyo Bento Kitchen: six restaurants and 48
available backend dishes, of which 41 photographed dishes appear in customer menus.
See [the current setup guide](README.md) and [Tokyo notes](TOKYO_READINESS.md).
No seed or database write is performed by this documentation update.

## Final catalogue

All prices below are integer GBP pence. Image paths are served by the React/Vite
frontend origin, not by Go. No remote runtime image requests are needed.

### Burger District

| Dish                        | Category | Pence | Available | Image               |
| --------------------------- | -------- | ----: | --------- | ------------------- |
| Classic Smash Burger        | Burgers  |   895 | Yes       | Photograph          |
| Smoky BBQ Bacon Burger      | Burgers  |  1095 | Yes       | Photograph          |
| Buttermilk Chicken Burger   | Burgers  |   945 | Yes       | Photograph          |
| Garden Bean Burger          | Burgers  |   825 | Yes       | Neutral placeholder |
| Sea Salt Skin-on Fries      | Sides    |   325 | Yes       | Photograph          |
| Loaded Cheese Fries         | Sides    |   495 | Yes       | Photograph          |
| Buffalo Wings (6 pieces)    | Wings    |   695 | Yes       | Photograph          |
| Sticky BBQ Wings (6 pieces) | Wings    |   695 | Yes       | Photograph          |

Cover: `/demo-images/photos/covers/burger-district.jpg`. Full dish descriptions and exact image paths are in `catalog.json`.

### Bella Italia Kitchen

| Dish                              | Category | Pence | Available | Image               |
| --------------------------------- | -------- | ----: | --------- | ------------------- |
| Margherita Pizza (12 inch)        | Pizzas   |   995 | Yes       | Photograph          |
| Pepperoni Pizza (12 inch)         | Pizzas   |  1195 | Yes       | Photograph          |
| Roasted Vegetable Pizza (12 inch) | Pizzas   |  1125 | Yes       | Photograph          |
| Spaghetti Bolognese               | Pasta    |  1095 | Yes       | Photograph          |
| Chicken Mushroom Tagliatelle      | Pasta    |  1245 | Yes       | Neutral placeholder |
| Penne Arrabbiata                  | Pasta    |   945 | Yes       | Photograph          |
| Cheesy Garlic Bread               | Sides    |   425 | Yes       | Photograph          |
| Classic Beef Lasagne              | Pasta    |  1195 | Yes       | Photograph          |

Cover: `/demo-images/photos/covers/bella-italia-kitchen.jpg`. Full dish descriptions and exact image paths are in `catalog.json`.

### Spice Route

| Dish                      | Category | Pence | Available | Image               |
| ------------------------- | -------- | ----: | --------- | ------------------- |
| Chicken Dum Biryani       | Biryani  |  1195 | Yes       | Photograph          |
| Vegetable Biryani         | Biryani  |   995 | Yes       | Photograph          |
| Butter Chicken            | Curries  |  1095 | Yes       | Photograph          |
| Chana Masala              | Curries  |   845 | Yes       | Photograph          |
| Palak Paneer              | Curries  |   945 | Yes       | Photograph          |
| Garlic and Coriander Naan | Breads   |   295 | Yes       | Neutral placeholder |
| Basmati Pilau Rice        | Sides    |   325 | Yes       | Photograph          |
| Lamb Rogan Josh           | Curries  |  1295 | Yes       | Neutral placeholder |

Cover: `/demo-images/photos/covers/spice-route.jpg`. Full dish descriptions and exact image paths are in `catalog.json`.

### The Shawarma House

| Dish                          | Category | Pence | Available | Image               |
| ----------------------------- | -------- | ----: | --------- | ------------------- |
| Chicken Shawarma Wrap         | Wraps    |   795 | Yes       | Photograph          |
| Lamb Shawarma Wrap            | Wraps    |   895 | Yes       | Neutral placeholder |
| Falafel and Hummus Wrap       | Wraps    |   695 | Yes       | Photograph          |
| Chicken Shawarma Rice Platter | Grills   |  1195 | Yes       | Photograph          |
| Lamb Kofta Grill              | Grills   |  1245 | Yes       | Neutral placeholder |
| Hummus and Flatbread          | Sides    |   395 | Yes       | Photograph          |
| Falafel Bites (5 pieces)      | Sides    |   445 | Yes       | Photograph          |
| House Mixed Grill             | Grills   |  1595 | Yes       | Neutral placeholder |

Cover: `/demo-images/photos/covers/shawarma-house.jpg`. Full dish descriptions and exact image paths are in `catalog.json`.

### Sweet Treats Café

| Dish                        | Category   | Pence | Available | Image      |
| --------------------------- | ---------- | ----: | --------- | ---------- |
| Strawberry Chocolate Waffle | Waffles    |   745 | Yes       | Photograph |
| Banoffee Waffle             | Waffles    |   725 | Yes       | Photograph |
| Warm Chocolate Brownie      | Cakes      |   495 | Yes       | Photograph |
| Strawberry Cheesecake Slice | Cakes      |   525 | Yes       | Photograph |
| Chocolate Fudge Cake Slice  | Cakes      |   495 | Yes       | Photograph |
| Vanilla Milkshake           | Milkshakes |   445 | Yes       | Photograph |
| Cookies and Cream Milkshake | Milkshakes |   525 | Yes       | Photograph |
| Red Velvet Cake Slice       | Cakes      |   545 | Yes       | Photograph |

Cover: `/demo-images/photos/covers/sweet-treats-cafe.jpg`. Full dish descriptions and exact image paths are in `catalog.json`.

## Historical preflight

The original preparation checked local target identity, permissions, foreign keys,
name/email collisions, categories, custom triggers/rules, RLS/partitions and sequences
before the first seed. Its counts and absent-registry findings predated local setup
and must not be used as today's database state. No private database inventory or
backup/receipt location is needed for tutor review.

For a fresh environment, perform those checks against its own approved target.
For an existing installation, inspect registry and data drift rather than assuming
an empty database or resetting seed metadata. Follow [README.md](README.md).

## Ordering and demo constraints

- Uses the existing authenticated `POST /api/cart/items`, `GET /api/cart`,
  `PUT /api/cart/items/{id}`, `GET /api/checkout/summary` and `POST /api/orders`.
- Public listing/details require active restaurants; every new seed restaurant
  is explicitly active. All 40 new items are available, positive-priced and
  connected to their actual restaurant/category IDs.
- All 40 catalogue mappings are checked offline against the real Go cart and
  checkout services, using test-only in-memory repository storage. No mock cart
  or alternate checkout is added to the application.
- Each cart can hold dishes from only one restaurant. Existing carts are not
  cleared. Use a customer with an empty cart or complete their current order.
- Checkout places a real order in the local database after delivery validation;
  it does not process a monetary payment. Backend totals remain authoritative.
- Existing drift/name collisions cause a rollback, not automatic repair. Older
  seed records with different availability/image values are never updated.

## Setup and validation status

Earlier preparation documents said “not executed” because execution had not yet
been authorised at that stage. The owner now reports that local FoodHub setup and
manager login are working. This review does not reconstruct insert counts, assert
an independently observed seed rerun, or test a fresh database installation.

Offline tests cover catalogue mappings, production cart/checkout contracts, image
integrity and environment/ownership guards. Use the validation commands in the
current setup guide; exclude the database-writing `TestOrderSnapshot`. Historical
checksums and test counts are superseded by the current files and test output.

All original 40 dishes remain available in the backend catalogue. Seven have blank
photo paths and are hidden from customer menus. Images that fail to load show a
neutral placeholder; archived illustrations are not rendered as fallbacks.

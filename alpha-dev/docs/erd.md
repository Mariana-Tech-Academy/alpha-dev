# FoodHub ERD

All tables also have `id`, `created_at`, `updated_at` and `deleted_at` (soft delete) from `gorm.Model`.
Money columns are integers in minor units (kobo/cents).

```mermaid
erDiagram
    USERS ||--o| CARTS : "has one"
    USERS ||--o{ ORDERS : "places"
    USERS ||--o{ RESTAURANTS : "manages"
    RESTAURANTS ||--o{ MENU_ITEMS : "has"
    RESTAURANTS ||--o{ ORDERS : "receives"
    CATEGORIES ||--o{ MENU_ITEMS : "groups"
    CARTS ||--o{ CART_ITEMS : "contains"
    MENU_ITEMS ||--o{ CART_ITEMS : "referenced by"
    ORDERS ||--|{ ORDER_ITEMS : "contains"
    MENU_ITEMS ||--o{ ORDER_ITEMS : "referenced by"

    USERS {
        uint id PK
        string first_name
        string last_name
        string email UK
        string phone
        string password "bcrypt hash"
        string role "customer | manager | admin"
        string status "active | suspended"
    }
    RESTAURANTS {
        uint id PK
        uint manager_id FK
        string name
        string description
        string address
        string city
        string state
        string phone
        string email
        string image_url
        string opening_time "HH:MM"
        string closing_time "HH:MM"
        string status "pending_approval | active | inactive | suspended"
    }
    CATEGORIES {
        uint id PK
        string name UK
    }
    MENU_ITEMS {
        uint id PK
        uint restaurant_id FK
        uint category_id FK
        string name
        string description
        int64 price
        string image_url
        bool available
    }
    CARTS {
        uint id PK
        uint user_id FK,UK
        uint restaurant_id FK "nullable; one restaurant per cart"
    }
    CART_ITEMS {
        uint id PK
        uint cart_id FK "unique with menu_item_id"
        uint menu_item_id FK
        int quantity "> 0"
    }
    ORDERS {
        uint id PK
        string order_number UK "ORD-YYYYMMDD-NNN"
        uint customer_id FK
        uint restaurant_id FK
        string status
        int64 subtotal
        int64 delivery_fee
        int64 total
        string delivery_name "snapshot"
        string delivery_phone "snapshot"
        string delivery_address "snapshot"
        string delivery_city "snapshot"
        string delivery_state "snapshot"
        string delivery_instructions "snapshot"
    }
    ORDER_ITEMS {
        uint id PK
        uint order_id FK
        uint menu_item_id FK
        string name "snapshot"
        int64 unit_price "snapshot"
        int quantity "> 0"
        int64 line_total
    }
```

## Decisions the team still needs to document

- **Cart conflict rule (PRD 4.4):** reject adding an item from a different restaurant, or ask the customer to clear the cart.
- **Cancellation rule (PRD 4.8):** which statuses allow cancellation and who may cancel.

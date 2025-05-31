# Custom Scalar Examples

This document provides examples of using the custom scalars implemented in the go-quickgraph-sample application.

## Registered Custom Scalars

The sample application demonstrates the following custom scalar types:

- **DateTime**: RFC3339 formatted date-time strings
- **EmployeeID**: Unique identifiers for employees
- **ProductID**: Unique identifiers for products
- **HexColor**: Hexadecimal color values (e.g., #FF0000)
- **Money**: Monetary amounts in cents
- **EmailAddress**: Validated email addresses
- **URL**: Valid URL strings

## Example Queries and Mutations

### Basic Scalar Usage

#### Email Validation
```graphql
query {
  validateEmail(email: "user@example.com")
}
```

Response:
```json
{
  "data": {
    "validateEmail": true
  }
}
```

#### Creating a Colored Product
```graphql
mutation {
  createColoredProduct(
    name: "Red Widget"
    price: 2500
    color: "#FF0000"
  ) {
    id
    name
    price
    color
  }
}
```

Response:
```json
{
  "data": {
    "createColoredProduct": {
      "id": "prod_1748700034",
      "name": "Red Widget", 
      "price": 2500,
      "color": "#FF0000"
    }
  }
}
```

### Employee ID Scalar

```graphql
query {
  getEmployeeByIDScalar(id: "emp_123") {
    ID
    Name
    Email
  }
}
```

Note: This will fail with current data since the employee storage uses integer IDs, but demonstrates the scalar parsing.

### Color Validation Examples

Valid colors:
- `#FF0000` (red)
- `#00FF00` (green) 
- `#0000FF` (blue)
- `#FFF` (short form white)

Invalid colors will be rejected:
- `FF0000` (missing #)
- `#GG0000` (invalid hex)
- `#12345` (wrong length)

### Money Scalar Examples

Money values are stored in cents:
- `1000` = $10.00
- `2500` = $25.00
- `99` = $0.99

```graphql
mutation {
  createColoredProduct(
    name: "Expensive Item"
    price: 999999
    color: "#FFD700"
  ) {
    price  # Returns: 999999 (which represents $9,999.99)
  }
}
```

### Email Address Validation

Valid emails:
- `user@example.com`
- `test.email+tag@domain.co.uk`
- `simple@domain.org`

Invalid emails will be rejected:
- `invalid-email` (no @)
- `@domain.com` (no local part)
- `user@` (no domain)
- `user@@domain.com` (multiple @)

## Schema Integration

All custom scalars appear in the generated GraphQL schema:

```graphql
scalar DateTime # RFC3339 formatted date-time string
scalar EmailAddress # Valid email address
scalar EmployeeID # Unique identifier for employees
scalar HexColor # Hexadecimal color representation (e.g., #FF0000)
scalar Money # Monetary amount in cents
scalar ProductID # Unique identifier for products
scalar URL # Valid URL
```

## Error Handling

Custom scalars include validation and provide clear error messages:

### Invalid Email
```graphql
query {
  validateEmail(email: "invalid-email")
}
```

This would fail during parsing with an error like:
```
"invalid email address: invalid-email"
```

### Invalid Color
```graphql
mutation {
  createColoredProduct(
    name: "Bad Color"
    price: 1000
    color: "not-a-color"
  ) {
    color
  }
}
```

This would fail with:
```
"invalid hex color format: not-a-color"
```

## Testing with CLI

You can test these scalars using the command-line interface:

```bash
# Test email validation
go run ./cmd/server -query 'query { validateEmail(email: "test@example.com") }'

# Test colored product creation
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Blue Widget", price: 1500, color: "#0000FF") { id name price color } }'

# Test with invalid color (will fail)
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Bad Widget", price: 1000, color: "invalid") { id } }'
```

## Using in GraphQL Clients

When using these scalars in GraphQL clients, remember:

1. **Variables**: Custom scalars are passed as their underlying JSON types
2. **Parsing**: Client-side parsing/validation may be needed
3. **TypeScript**: Generate appropriate type definitions for custom scalars

Example with variables:
```graphql
mutation CreateProduct($name: String!, $price: Money!, $color: HexColor!) {
  createColoredProduct(name: $name, price: $price, color: $color) {
    id
    name
    price
    color
  }
}
```

Variables:
```json
{
  "name": "Green Widget",
  "price": 1999,
  "color": "#00FF00"
}
```

## Implementation Notes

- Custom scalars are registered before any handlers to ensure proper type resolution
- Validation occurs during both input parsing and literal parsing
- Error messages include context about the expected format
- All scalars support both value and pointer types for flexibility
- The implementation includes helper functions for common operations (e.g., Money formatting)
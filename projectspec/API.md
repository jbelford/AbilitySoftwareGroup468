# REST API Specification

Example error response:
```json
{
  "success": false,
  "message": "Not enough money in account"
}
```


#### `POST /{user_id}/add?amount={amount}`
```json
{
  "success": true
}
```

#### `GET /{user_id}/quote?stock={symbol}`

```json
{
  "success": true,
  "stock": "ABC",
  "quote": 12.50
}
```

#### `POST /{user_id}/buy?stock={symbol}&amount={amount}`

```json
{
  "success": true,
  "amount_requested": 200,
  "real_amount": 195.15,
  "shares": 20,
  "expiration": 1516767552619
}
```

#### `POST /{user_id}/commit_buy`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 20,
  "paid": 195.15
}
```

#### `POST /{user_id}/cancel_buy`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 20
}
```

#### `POST /{user_id}/sell?stock={symbol}&amount={amount}`

```json
{
  "success": true,
  "amount_requested": 200,
  "real_amount": 182.21,
  "shares": 19,
  "expiration": 1516767552619
}
```

#### `POST /{user_id}/commit_sell`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 19,
  "received": 182.21
}
```

#### `POST /{user_id}/cancel_sell`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 19
}
```

#### `POST /{user_id}/set_buy_amount?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `POST /{user_id}/cancel_set_buy?stock={symbol}`

```json
{
  "success": true,
  "stock": "ABC"
}
```

#### `POST /{user_id}/set_buy_trigger?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `POST /{user_id}/set_sell_amount?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `POST /{user_id}/set_sell_trigger?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `POST /{user_id}/cancel_set_sell?stock={symbol}`

```json
{
  "success": true
}
```

#### `POST /{user_id}/dumplog?filename={filename}`

```json
{
  "success": true
}
```

#### `POST /{admin_id}/dumplog?filename={filename}`

```json
{
  "success": true
}
```

#### `GET /{user_id}/display_summary`

```json
{
  "success": true,
  "status": {
    "balance": 2000
  },
  "transactions": [
    {
      "type": "BUY",
      "triggered": false,
      "stock": "ABC",
      "amount": 192.15,
      "shares": 20,
      "timestamp": 1516767552619
    }
  ],
  "triggers": [
    {
      "stock": "ABC",
      "type": "SELL",
      "amount": 200,
      "when": 10.50
    }
  ]
}
```

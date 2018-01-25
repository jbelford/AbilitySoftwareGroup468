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
  "quote": 1250
}
```

#### `POST /{user_id}/buy?stock={symbol}&amount={amount}`

```json
{
  "success": true,
  "amount_requested": 20000,
  "real_amount": 19515,
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
  "paid": 19515
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
  "amount_requested": 20000,
  "real_amount": 18221,
  "shares": 19,
  "shares_afford": 10,
  "afford_amount": 13000,
  "expiration": 1516767552619
}
```

#### `POST /{user_id}/commit_sell`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 19,
  "received": 18221
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
      "amount": 19215,
      "shares": 20,
      "timestamp": 1516767552619
    }
  ],
  "triggers": [
    {
      "stock": "ABC",
      "type": "SELL",
      "amount": 200,
      "when": 1050
    }
  ]
}
```

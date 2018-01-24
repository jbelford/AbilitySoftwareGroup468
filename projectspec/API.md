# REST API Specification

Example error response:
```json
{
  "success": false,
  "message": "Not enough money in account"
}
```


#### `/{user_id}/add?amount={amount}`
```json
{
  "success": true
}
```

#### `/{user_id}/quote?stock={symbol}`

```json
{
  "success": true,
  "stock": "ABC",
  "quote": 12.50
}
```

#### `/{user_id}/buy?stock={symbol}&amount={amount}`

```json
{
  "success": true,
  "amount_requested": 200,
  "real_amount": 195.15,
  "shares": 20,
  "expiration": 1516767552619
}
```

#### `/{user_id}/commit_buy`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 20,
  "paid": 195.15
}
```

#### `/{user_id}/cancel_buy`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 20
}
```

#### `/{user_id}/sell?stock={symbol}&amount={amount}`

```json
{
  "success": true,
  "amount_requested": 200,
  "real_amount": 182.21,
  "shares": 19,
  "expiration": 1516767552619
}
```

#### `/{user_id}/commit_sell`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 19,
  "received": 182.21
}
```

#### `/{user_id}/cancel_sell`

```json
{
  "success": true,
  "stock": "ABC",
  "shares": 19
}
```

#### `/{user_id}/set_buy_amount?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `/{user_id}/cancel_set_buy?stock={symbol}`

```json
{
  "success": true,
  "stock": "ABC"
}
```

#### `/{user_id}/set_buy_trigger?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `/{user_id}/set_sell_amount?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `/{user_id}/set_sell_trigger?stock={symbol}&amount={amount}`

```json
{
  "success": true
}
```

#### `/{user_id}/cancel_set_sell?stock={symbol}`

```json
{
  "success": true
}
```

#### `/{user_id}/dumplog?filename={filename}`

```json
{
  "success": true
}
```

#### `/{admin_id}/dumplog?filename={filename}`

```json
{
  "success": true
}
```

#### `/{user_id}/display_summary`

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

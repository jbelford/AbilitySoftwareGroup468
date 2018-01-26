# Database Schema

## Collections

#### Users

```ts
{
    "userId": string;
    "balance": number;
    "stock": { [symbol:string]: number; }
    "triggers": ObjectId[];
}
```

#### Triggers

```ts
{
    "_id": ObjectId;
    "userId": string;
    "stock": string;
    "type": "BUY" | "SELL";
    "amount": number;
    "when": number;
}
```

#### Transactions

```ts
{
    "type": "BUY" | "SELL";
    "triggered": bool;
    "stock": string;
    "amount": number;
    "shares": number;
    "timestamp": number;
}
```
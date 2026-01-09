# Card Testing Detection

Demonstrates detection of **card testing** - a common fraud pattern where fraudsters test stolen credit cards with small transactions.

## The Attack Pattern

Fraudsters obtain lists of stolen credit card numbers and need to verify which ones are still active. They:

1. Make many small transactions ($1-5) to avoid triggering amount-based alerts
2. Use different cards for each transaction (testing multiple cards quickly)
3. Execute rapidly (10-20 transactions in 5 minutes)
4. Often use the same merchant/account

## Detection Strategy

This example uses 4 key features to detect card testing:

### 1. **Velocity** (events per minute)
- Normal users: 0.1-0.5 tx/min
- Card testing: >2 tx/min
- **Rule**: `velocity > 2.0` → +30 risk points

### 2. **Transaction Count** (5-minute window)
- Normal users: 1-3 transactions
- Card testing: 10-20 transactions
- **Rule**: `count >= 10` → +25 risk points

### 3. **Unique Cards Ratio** (unique/total)
- Normal users: 0.0-0.3 (usually same card)
- Card testing: 0.8-1.0 (different card each time)
- **Rule**: `ratio > 0.8` → +30 risk points

### 4. **Average Amount**
- Normal users: $20-100
- Card testing: $1-5
- **Rule**: `avg < $10` → +15 risk points

## Risk Scoring

```
Total Risk Score = sum of triggered rules

< 40:  ✅ Normal activity
40-59: ⚠️  Suspicious - flag for review
≥ 60:  🚨 Card testing detected - BLOCK
```

## Running the Example

```bash
cd examples/card-testing
go run main.go
```

## Expected Output

### Normal User (Risk Score: 0)
```
Transactions (5 min):     3
Velocity:                 0.6 tx/min
Distinct cards:           1
Unique cards ratio:       0.33
Average amount:           $49.16
Max amount:               $89.00

Risk Score: 0/100
✅ Normal activity
```

### Card Testing Attack (Risk Score: 100)
```
Transactions (5 min):     15
Velocity:                 3.0 tx/min
Distinct cards:           15
Unique cards ratio:       1.00
Average amount:           $3.25
Max amount:               $4.87

Risk Score: 100/100
Risk Factors:
  • high velocity (3.0 tx/min)
  • high transaction count (15 in 5 min)
  • high card diversity (100% unique)
  • small amounts (avg $3.25)

🚨 CARD TESTING DETECTED - BLOCK IMMEDIATELY
```

## Real-World Tuning

These thresholds work for most e-commerce scenarios, but you may need to adjust based on your business:

- **Digital goods**: Lower velocity threshold (2.0 → 1.5)
- **Subscription services**: Higher unique ratio tolerance (0.8 → 0.9)
- **High-value merchants**: Higher amount threshold ($10 → $20)

## Prevention Measures

Once card testing is detected:

1. **Immediate**: Block the account/IP
2. **Short-term**: Add CAPTCHA for new cards
3. **Long-term**: Require 3D Secure for first transaction

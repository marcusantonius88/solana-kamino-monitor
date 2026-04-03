# 📊 Solana Kamino Health Factor Monitor

A lightweight Go REST API to monitor Kamino Lending positions on Solana.

The project focuses on one key signal: `health factor`, the main metric used to understand how safe or risky a borrowing position is.

It is built for monitoring, not simulation.

---

## 🚀 Features

* Query a Solana wallet and inspect its Kamino lending obligations
* Calculate the `healthFactor` for the riskiest obligation in the wallet
* Validate Solana wallet addresses before processing requests
* Discover Kamino obligations directly from Solana RPC
* Return a simple API response that is easy to integrate into dashboards, bots, or alerts

---

## 🧰 Requirements

* Go 1.24+
* A Solana RPC endpoint

---

## ⚙️ Installation

### 1. Clone the repository

```bash
git clone <your-repository-url>
cd kamino-simulator
```

---

### 2. Install dependencies

```bash
go mod tidy
```

---

### 3. Configure environment variables

Create or update the `.env` file in the project root:

```env
PORT=8080
SOLANA_RPC_URL=https://api.mainnet-beta.solana.com
```

You can also use a custom RPC provider if you want better reliability or higher rate limits.

---

## ▶️ Running the API

Start the server with:

```bash
go run ./cmd/api
```

The application automatically loads variables from `.env` on startup.

You can also override them directly in the command:

```bash
PORT=8080 SOLANA_RPC_URL="https://your-rpc-endpoint" go run ./cmd/api
```

---

## 🔌 API Usage

### 📌 Get Wallet Position Health

```bash
curl "http://localhost:8080/positions/So11111111111111111111111111111111111111112"
```

Example response when the wallet has an active borrowing position:

```json
{
  "healthFactor": 1.42
}
```

If the wallet has no obligations with active debt:

```json
{}
```

---

## 🐞 Debug Mode

For discovery diagnostics, you can enable debug mode:

```bash
curl "http://localhost:8080/positions/So11111111111111111111111111111111111111112?debug=true"
```

In normal mode, the API uses a lean discovery flow to reduce the chance of `429` responses from the RPC provider.

With `debug=true`, it performs a broader scan to help investigate discovery issues.

---

## ⚠️ Error Responses

Invalid wallet:

```json
{
  "error": "invalid Solana wallet address: ..."
}
```

Internal error:

```json
{
  "error": "internal server error"
}
```

---

## 🏗️ Project Structure

```text
.
├── cmd/api
├── internal
│   ├── handler
│   ├── repository
│   └── service
└── pkg
    └── solana
```

The project follows a Clean Architecture-inspired structure to keep HTTP handling, business logic, and Solana-specific integrations separated.

---

## ✅ Current Scope

This MVP already includes:

* HTTP API endpoint for wallet monitoring
* Solana wallet validation
* Kamino obligation discovery
* Binary parsing of obligation data
* `healthFactor` calculation
* Basic logging

---

## 📌 Future Improvements

* Return additional metadata such as obligation count and obligation addresses
* Improve number formatting and precision
* Add support for alerting and external integrations
* Expose more monitoring-focused metrics beyond `health factor`

---

## 🤝 Contributing

Issues and pull requests are welcome.

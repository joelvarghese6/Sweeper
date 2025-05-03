# Wallet Protection API

A Go-based API for detecting various blockchain-based attacks on wallet addresses, including dust attacks and address poisoning.

## 🔍 Features

- **Dust Attack Detection**: Check if an address has been targeted by dust attacks
- **Address Poisoning Detection**: Identify address poisoning attempts
- **Transaction Filtering**: Filter out spam and attack transactions for a given address

## 🚀 Getting Started

### Prerequisites

- Go 1.16 or higher
- Git

### Installation

1. Clone the repository
```bash
git clone https://github.com/joelvarghese6/Sweeper.git
cd Sweeper
```

2. Install dependencies
```bash
go mod download
```

## 💻 Usage

### Running the API

Start the API server with:

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080` by default.

### API Endpoints

#### 1. Check for Dust Attacks

Retrieves all dust attacks that occurred on a specific address.

```
GET /api/check-dusted?Publickey={wallet_address}
```

**Parameters:**
- `Publickey` (required): The public key/address to check for dust attacks

**Example Response:**
```json
{
  "Code": "200",
  "Items": [
    {
      "Other": "2tBUH...",
      "Amount": "20.00",
      "Dust": "true",
      "Multi": "true",
      "Sig": "0xabc...",
    }
  ],
}
```

#### 2. Check for Address Poisoning

Checks if the address has been targeted by address poisoning attacks.

```
GET /api/check-address-poisoning?Publickey={wallet_address}
```

**Parameters:**
- `Publickey` (required): The public key/address to check for poisoning attempts

**Example Response:**
```json
{
  "Code": "200",
  "poisoning_attempts": [
    {
      "from_address": "0xdef...",
      "similiar_address": "2tBU...",
      "signature": "0x124...",
      "amount": "20.00"
    }
  ],
  "Poisoned": "true"
}
```

#### 3. Filter Transactions

Filters out spam and attack transactions for a given address.

```
GET /api/filter-transactions?Publickey={wallet_address}
```

**Parameters:**
- `Publickey` (required): The public key/address to filter transactions for

**Example Response:**
```json
{
  "address": "0x123...",
  "legitimate_transactions": [
    {
      "transaction_id": "0xghi...",
      "timestamp": "2023-10-14T11:20:00Z",
      "amount": "0.5",
      "type": "transfer"
    }
  ],
  "filtered_transactions": [
    {
      "transaction_id": "0xjkl...",
      "timestamp": "2023-10-15T14:30:00Z",
      "type": "dust_attack"
    }
  ],
  "total_legitimate": 1,
  "total_filtered": 1
}
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.


## 📧 Contact

Joel Varghese - [@joelvarghese_](https://twitter.com/joelvarghese_) - varghesejoel6@gmail.com

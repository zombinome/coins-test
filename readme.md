# Test excercise for Coins.ph

This excercise contains small wallet-like web service that allows you to see list of accounts,
list money transfer operations for each account and transfer money between accounts.

## Setup
Please checkout this repository from github.
In folder `deploy` you can find file `create_db.sql` which contains SQL script that can create posgtres database with all tables. Please run it on you posgres server.

Application code is located on `src` folder. Before build and run application, please open fule `src/.env` and update database connection string and port.

Build and run application from `src` folder.

## Development notes
As test excercise, this project is very limited by functionality. A lot of stuff was omitted to reduce time on implementing functionality and writing tests. For example, list paging and sorting is not supported.

### Stack
As per assignment, application is build using go-kit as web framework and postgres to store data.
Database connection string is loaded from .env file using `godotenv`. To query postrgres, `pgx` library is used. To mock work with DB in tests, `sqlmocks` is used.

### Database
Database creation script is located in `deploy` folder of repository.
Database contains 2 tables, `accounts` and `transfers`. 

`Accounts` table contains account number and current balance.

`Transfers` table contains amount of data transfered, source and dest accounts and unique transfer id. Transfer id is GUID and should be always provided by client to avoid double transfer in case if client decided to repeat same request to service for some reason.

Account balances and transfer amounts is stored as integer number in smallest denomination of currency. For example, for USD$ it would be cents, $ 12.50 would be stored as 1250. Service always expects tansfer amounts in same integer format. Please note, as there is only one currency, backend does not store or return currency name.

Protection against concurrency problems with money transfer is implemented using via locking affected rows in accounts until transaction ends (using SELECT ... FROM public.accounts ... FOR UPDATE query). All transactions has rollback on timeout, to avoid blocking DB records forever. Default trasnaction timeout is set to 5 seconds, which is arbitrary value.

### Architecture
Application is implemented as 2 busness services - AccountService (src/account) and TransferService (src/transfer). Additionally infrastructure code added to unify error handling and database interaction (src/errors and src/db).

Work with database wrapped in DbContext contract to simplify mocking services when writing tests and reduce amount of code repetition. DbContext has 2 implementations - pgxDbContext used to work with postgres (via pgx library) and mockDbContext is used in tests.

## API
Application created with RESTful architecture in mind. Application supports following requests:
* `GET /api/v1/accounts` - returns list of accounts
* `GET /api/v1/accounts/{accountNumber}/transfers` - returns list of money transfers for specific account
* `POST /api/v1/transfers` - transfers money between 2 accounts 



### List of accounts
`GET /api/v1/accounts`
Returns list ofaccounts in wallet.
Result format:
```
{
    "accounts": [
        {
            "number": 1,
            "balance": 1000
        },
        {
            "number": 2,
            "balance": 2000
        }
    ]
}
```

In case of error, request will return response code 500 and message error in body:
```
{
    "error": "error message"
}
```

### List of money transfers for account (history)
`GET /api/v1/accounts/{accountNumber}/transfers`
Returns list of money transfer operations for account with number `{accountNumber}`.
Returns:
```
{
    "transfers": [
        {
            "id": "9411b3e7-4f06-42d9-a1de-b996daf868f6",
            "account": 1,
            "toAccount": 2,
            "amount": 150,
            "direction": "outgoing",
            "createdAt": "2021-12-F17T21:31:00.643Z"
        },
        {
            "id": "ac5ed528-3fc7-44cd-9b39-795959781afa",
            "account": 1,
            "fromAccount": 2,
            "amount": 50,
            "direction": "incoming",
            "createdAt": "2021-12-F17T21:31:00.643Z"
        }
    ]
}
```
In case of error, request will return response 400 if account number is invalid, or 500 if there is some database error. Reposnse body would be like this:
```
{
    "error": "error message"
}
```

### Trasnfer money to another account
`POST /api/v1/transfers`
Transfers money to another account.
Request body has foloowing required fields:
```
{
    "id": "dc4214f0-6c39-4663-b43b-2ddcf72cee4e",
    "source": 1,
    "dest": 2,
    "amount": 150
}
```
`amount` should be positive integer number.

If money transfer successfully, you will get empty response with code 200.
If source, dest is missing or refers to not existing account, you will get error response with code 400.
If there is already exists transfer with same transfer id, it will return error with code 400.
If transfer amount is greated that source account balance, server will return error with code 400.
Other errors will produce response with code 500.

If error occured, response body would look like this:
```
{
    "error": "error message"
}
```

## Tests
I tried to cover main cases for services with unit tests. Integration tests is not there, as it is separate beast to tame (did not have enough time to learn and implement properly in go).

## Docker
Docker support is missing. I honestly had minimum experience with docker before, and I won't be able to do it quickly right now (docker compose + volumes + build image script).
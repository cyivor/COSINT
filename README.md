# COSINT
Collective Open-Source Intelligence (COSINT) is a tool which encompasses multiple OSINT tools and toolkits, keeping the data and automating searches.

## Before running
<sup>It is important you read this or it won't work</sup>

1. Your `.env`:
```
JWT_SECRET=your_jwt_secret
DB_KEY=your_db_key
```

2. Your `cosint.db`
You need to manually create the cosint userdatabase at this moment in time - I'll probably edit this readme at some point and realise what I'm typing right now isn't useful

## Todo
* short-term
  - [x] create user database
  - [x] create register page
  - [x] create API endpoint for snusbase
    - [ ] add local ratelimit 
    - [ ] search all fields unless field is specified


* long-term
  - [ ] add support for nosint
  - [ ] add support for maigret
  - [ ] add support for telegram bots
    * parsing modules will be needed
  - [ ] add dorking support
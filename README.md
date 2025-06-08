# COSINT
Collective Open-Source Intelligence (COSINT) is a tool which encompasses multiple OSINT tools and toolkits, keeping the data and automating searches.

## Notes
- Storage efficient
- Good security
- Rotating API routes (upon run-time)
- Quick
- Encrypted user database (with hashed passwords)

## Before running
<sup>It is important you read this or it won't work</sup>

1. Your `.env`:
```
JWT_SECRET=your_jwt_secret
DB_KEY=your_db_key
```

2. Your `cosint.db`
You need to manually create the cosint userdatabase at this moment in time - I'll probably edit this readme at some point and realise what I'm typing right now isn't useful

3. Ensure Maigret is installed via `pip` so you can use it globally
You can test by typing `maigret --version` into your terminal, if there's an output you're good to go. *If maigret was installed in a virtual environment, just source the activation file before using COSINT*

## Todo
* short-term
  - [x] create user database
  - [x] create register page
  - [x] create API endpoint for snusbase
    - [x] add local ratelimit 
    - [ ] search all fields unless field is specified


* long-term
  - [ ] add support for nosint
  - [ ] add support for maigret
    - [x] install maigret locally 
    - [ ] fix report generation
      * using virtual environment
  - [ ] add support for telegram bots
    * parsing modules will be needed
  - [ ] add dorking support

## Check developments
- [Dev branch](https://github.com/cyivor/COSINT/tree/dev)
- [Dev commits](https://github.com/cyivor/COSINT/commits/dev/)
- [Main branch](https://github.com/cyivor/COSINT)
- [Main commits](https://github.com/cyivor/COSINT/commits/main/)
## journal cli
### a little cli that records journals and provides nice views of past journals

features and todo lists
- encryption
    - [x] encrypt and decrypt entries

- writing entries
    - [x] multiline place to write
    - [x] option for title - if none provided, just use date
    - [x] tagging
    - [ ] option to throw in different media
- viewing entries
    - [ ] be able to load in past entry to edit
    - [ ] edit..history?
    - [ ] calendar view
    - [x] be able to search through entries based on tags | date

- misc.
    - [ ] be able to change text color, 
    - [ ] have a default program size (!!)
    - [ ] implement loading once io is in cmds
    - [ ] add in a bunch of different styles

- non-ui
    - [ ] organize encryption tasks into bubble tea cmds
    - [ ] make modules more cleanly nested - seperate different actions
    - [ ] make sure data is stored in memory when unencrypted AND whenever changed. as little decrypting as possible

## bugs
- [ ] readui doesn't store data away and reload it from memory once viewed for a second time
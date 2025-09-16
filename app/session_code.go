package app

//Function that generates a session code - use crypto/rand to make secure (no one brute forcing into session)
//Must be unique
//Must be easily readable (6 - 8 chars?)
//Create invite link with code to send to users
// Extract code from invite link to enter session

//Func:

//Generate random code
//Check if already exists - if does retry - (GetSessionByCode)

# simple-polling-web-app

## A simple polling web application

This is the source code for a simple web polling application built using [Go](https://golang.org/), [SQLite](https://sqlite.org/index.html), and HTML (no JavaScript!)

### What you can do

#### As a Poll Administrator:

- sign up for the website using an email address and password
  - **N.B.** This site is not secure, so do _not_ use a password you would mind being published on a public website. I recommend something simple like `123`.
- create polls including a title and
- edit polls
- view poll results

#### As a Poll Responder:

- visit the front page of this application and see a list of all the polls
- click on a poll, answer it and submit your response
- respond to a poll without logging in

### What you can't (yet) do

- access the website securely
- scales, multiple choice, or free response questions
- delete polls
- delete questions
- identify responders (by IP or any other method)
- edit responses
- correlate responses
- aggregate or analyze responses
- report when it is down

## What I would do if I had more time:

Rather than have the UI be completely static HTML web forms, I would use a framework like [Apollo](https://www.apollographql.com/) to manage user state dynamically on the client-side and manage communication between the client and server using GraphQL. FWIW, I began developing such a backend server [here](https://github.com/MadDataScience/simple-poll-api), and the frontend in the branch: [gatsby-apollo](https://github.com/MadDataScience/simple-polling-web-app/tree/gatsby-apollo).

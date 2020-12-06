import React from "react"
import { Link, graphql } from "gatsby"
import { getUser, isLoggedIn } from "../services/auth"

import Layout from "../components/layout"

export default function Home({ data }) {
  console.log(data)
  return (
    <Layout>
      <h1>Hello {isLoggedIn() ? getUser().name : "world"}!</h1>
      <p>
        {isLoggedIn() ? (
          <>
            You are logged in, so check your{" "}
            <Link to="/app/profile">profile</Link>
          </>
        ) : (
          <>
            You should <Link to="/app/login">log in</Link> to see restricted
            content
          </>
        )}
      </p>
      <h4>{data.backend.polls.length} Polls</h4>
      {data.backend.polls.map(poll => (
        <div key={poll.pollID}>
          <h3>{poll.title}</h3>
        </div>
      ))}
    </Layout>
  )
}

export const query = graphql`
  query {
    backend {
      polls {
        pollID
        title
      }
    }
  }
`

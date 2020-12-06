/**
 * Configure your Gatsby site with this file.
 *
 * See: https://www.gatsbyjs.com/docs/gatsby-config/
 */

module.exports = {
  /* Your site config here */
  plugins: [
    {
      resolve: "gatsby-source-graphql",
      options: {
        // Arbitrary name for the remote schema Query type
        typeName: "BackEnd",
        // Field under which the remote schema will be accessible. You'll use this in your Gatsby query
        fieldName: "backend",
        // Url to query from
        url: "http://localhost:5051/query",
      },
    }
  ],
}

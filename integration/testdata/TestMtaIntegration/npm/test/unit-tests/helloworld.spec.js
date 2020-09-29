const hello = require('../../srv/hello.js')

describe("hello world route", () => {
  it("responds with \"Hello, World!\"", () => {

    hello.sum(1,2)
    hello.multiply(1,1)
  });
});

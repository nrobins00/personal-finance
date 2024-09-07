let plaidLinkHandler;
const linkTokenResp = fetch("http://localhost:8080/api/linktoken", {
  method: "POST",
})
  .then((resp) => resp.json())
  .then((data) => {
    console.log(data);
    const linkToken = data.link_token;
    plaidLinkHandler = Plaid.create({
      token: linkToken,
      onSuccess: (public_token, metadata) => {
        exchangePublicToken(public_token, metadata);
      },
      onLoad: () => {},
      onExit: (err, metadata) => {},
      onEvent: (eventName, metadata) => {},
    });
  });

async function exchangePublicToken(public_token, metadata) {
  const startOfUserId = window.location.pathname.lastIndexOf("/") + 1;
  const userId = window.location.pathname.substring(startOfUserId);
  const hostname = window.location.host;
  const url = "http://" + hostname + "/api/publicToken/" + userId;
  const response = await fetch(url, {
    // TODO: figure out how to handle IP
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      //UserId: userId,
    },
    body: JSON.stringify({ Public_token: public_token }),
  });
  const data = await response.json();
  console.log(data.access_token);
}

function linkClickHandler() {
  console.log(plaidLinkHandler);
  plaidLinkHandler.open();
}

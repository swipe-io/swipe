import ClientServiceInterface from "./../transport/jsonrpc/jsclient";

var c = new ClientServiceInterface("http://localhost:9000");

c.getAll().then((data) => {});

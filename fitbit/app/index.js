import document from "document";
import * as messaging from "messaging";

let statusText = document.getElementById("status");
statusText.text = "Waiting...";

messaging.peerSocket.addEventListener("open", (evt) => {
  messaging.peerSocket.send({
    "url": "https://ifconfig.me",
  });
});

messaging.peerSocket.addEventListener("message", (evt) => {
  statusText.text = JSON.stringify(evt.data);
});

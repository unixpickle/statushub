import document from "document";
import * as messaging from "messaging";

const resultBox = document.getElementById('result-box');
resultBox.text = 'Tap refresh.';

const refreshButton = document.getElementById('refresh-button');
refreshButton.addEventListener('click', (evt) => {
  if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
    resultBox.text = 'Loading...';
    messaging.peerSocket.send({});
  } else {
    resultBox.text = 'Not connected to peer.';
  }
});

messaging.peerSocket.addEventListener("open", (evt) => {
  messaging.peerSocket.send({});
});

messaging.peerSocket.addEventListener("message", (evt) => {
  if (evt.data['error']) {
    resultBox.text = evt.data['error'];
  } else {
    resultBox.text = evt.data['data'];
  }
});

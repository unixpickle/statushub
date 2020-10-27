import document from "document";
import * as messaging from "messaging";

const resultBoxes = document.getElementsByClassName('result-box');
const resultBoxTitles = document.getElementsByClassName('result-box-title');
const refreshButton = document.getElementById('refresh-button');

function clearResults() {
  resultBoxes.forEach((x) => x.text = '');
  resultBoxTitles.forEach((x) => x.text = '');
}

function showResults(results) {
  clearResults();
  results.forEach((x, i) => {
    if (i < resultBoxes.length) {
      resultBoxTitles[i].text = x[0];
      resultBoxes[i].text = x[1];
    }
  });
}

function showMessage(msg) {
  clearResults();
  resultBoxes[0].text = msg;
}

function initialize() {
  showMessage('Not connected to peer.');

  refreshButton.addEventListener('click', (evt) => {
    if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
      showMessage('Loading...');
      messaging.peerSocket.send({});
    } else {
      showMessage('Not connected to peer.');
    }
  });

  messaging.peerSocket.addEventListener("open", (evt) => {
    messaging.peerSocket.send({});
  });

  messaging.peerSocket.addEventListener("message", (evt) => {
    if (evt.data['error']) {
      showMessage(evt.data['error']);
    } else {
      showResults(evt.data['data']);
    }
  });
}

initialize();

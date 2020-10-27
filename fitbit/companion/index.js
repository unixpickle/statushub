import * as messaging from "messaging";
import { settingsStorage } from "settings";

messaging.peerSocket.addEventListener("message", (evt) => {
  makeRequest();
});

function getBaseURL() {
  return JSON.parse(settingsStorage.getItem('baseurl') || '{"name":""}')["name"];
}

function getPassword() {
  return JSON.parse(settingsStorage.getItem('password') || '{"name":""}')["name"];
}

function makeRequest() {
  const url = getBaseURL() + '/api/overview?password=' + encodeURIComponent(getPassword())
  fetch(url, {
        credentials: 'include',
        redirect: 'manual',
  }).then((x) => {
    return x.json();
  }).then((data) => {
    if (data['error']) {
      messaging.peerSocket.send({error: data['error']});
    } else {
      let rows = [];
      data['data'].forEach((x) => {
        rows.push([x['serviceName'], x['message']]);
      });
      messaging.peerSocket.send({data: rows});
    }
  }).catch((err) => {
    messaging.peerSocket.send({error: 'Login error: ' + err});
  });
}

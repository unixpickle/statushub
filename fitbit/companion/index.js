import * as messaging from "messaging";
import { settingsStorage } from "settings";

const MAX_ROWS = 13;

messaging.peerSocket.addEventListener("message", (evt) => {
  if (evt.data['service']) {
    serviceLogRequest(evt.data['service']);    
  } else {
    overviewRequest();
  }
});

function getBaseURL() {
  return JSON.parse(settingsStorage.getItem('baseurl') || '{"name":""}')["name"];
}

function getPassword() {
  return JSON.parse(settingsStorage.getItem('password') || '{"name":""}')["name"];
}

function overviewRequest() {
  const url = getBaseURL() + '/api/overview?password=' + encodeURIComponent(getPassword())
  fetch(url, {
    credentials: 'include',
    redirect: 'manual',
  }).then((x) => {
    return x.json();
  }).then((data) => {
    if (data['error']) {
      messaging.peerSocket.send({service: null, error: data['error']});
    } else {
      const rows = data['data'].map((x) => [x['serviceName'], x['message']]);
      messaging.peerSocket.send({service: null, data: rows});
    }
  }).catch((err) => {
    messaging.peerSocket.send({service: null, error: 'Request error: ' + err});
  });
}

function serviceLogRequest(serviceName) {
  const url = getBaseURL() + '/api/serviceLog?password=' + encodeURIComponent(getPassword())
  fetch(url, {
    credentials: 'include',
    redirect: 'manual',
    method: 'POST',
    body: JSON.stringify({service: serviceName}),
  }).then((x) => {
    return x.json();
  }).then((data) => {
    if (data['error']) {
      messaging.peerSocket.send({service: serviceName, error: data['error']});
    } else {
      const rows = data['data'].slice(0, MAX_ROWS).map((x) => x['message']);
      messaging.peerSocket.send({service: serviceName, data: rows});
    }
  }).catch((err) => {
    messaging.peerSocket.send({service: serviceName, error: 'Request error: ' + err});
  });
}

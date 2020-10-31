import * as messaging from "messaging";
import { settingsStorage } from "settings";

messaging.peerSocket.addEventListener("message", (evt) => {
  if (evt['service']) {
    serviceLogRequest(evt['service']);    
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
      let rows = [];
      data['data'].forEach((x) => {
        rows.push([x['serviceName'], x['message']]);
      });
      messaging.peerSocket.send({service: null, data: rows});
    }
  }).catch((err) => {
    messaging.peerSocket.send({service: null, error: 'Login error: ' + err});
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
      let rows = [];
      data['data'].forEach((x) => {
        rows.push(x['message']);
      });
      messaging.peerSocket.send({service: serviceName, data: rows});
    }
  }).catch((err) => {
    messaging.peerSocket.send({service: serviceName, error: 'Login error: ' + err});
  });
}

import * as messaging from "messaging";
import { settingsStorage } from "settings";

messaging.peerSocket.addEventListener("message", (evt) => {
  const requestURL = evt.data['url'];
  fetch(requestURL).then((x) => x.text()).then((x) => {
    const idx = x.indexOf('<strong id="ip_address">') || 0;
    let ipAddr = x.substr(idx).split('>')[1].split('<')[0];
    ipAddr += '/' + getBaseURL() + '/' + getPassword();
    if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
      messaging.peerSocket.send(ipAddr);
    }
  });
});

function getBaseURL() {
  return JSON.parse(settingsStorage.getItem('baseurl') || '{"name":""}')["name"];
}

function getPassword() {
  return JSON.parse(settingsStorage.getItem('password') || '{"name":""}')["name"];
}

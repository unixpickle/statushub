import * as messaging from "messaging";

messaging.peerSocket.addEventListener("message", (evt) => {
  const requestURL = evt.data['url'];
  fetch(requestURL).then((x) => x.text()).then((x) => {
    const idx = x.indexOf('<strong id="ip_address">') || 0;
    const ipAddr = x.substr(idx).split('>')[1].split('<')[0];
    if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
      messaging.peerSocket.send(ipAddr);
    }
  });
});

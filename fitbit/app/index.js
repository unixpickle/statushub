import document from "document";
import * as messaging from "messaging";

class OverviewPage {
  constructor() {
    this.resultBoxes = document.getElementsByClassName('result-box');
    this.resultBoxTitles = document.getElementsByClassName('result-box-title');
    this.refreshButton = document.getElementById('refresh-button');
    
    this.showMessage('Not connected to peer.');

    this.refreshButton.addEventListener('click', () => this.refresh());
  }
  
  clearResults() {
    this.resultBoxes.forEach((x) => x.text = '');
    this.resultBoxTitles.forEach((x) => x.text = '');
  }
  
  showResults(results) {
    this.clearResults();
    results.forEach((x, i) => {
      if (i < this.resultBoxes.length) {
        this.resultBoxTitles[i].text = x[0];
        this.resultBoxes[i].text = x[1];
      }
    });
  }
  
  showMessage(msg) {
    this.clearResults();
    this.resultBoxes[0].text = msg;
  }
  
  refresh() {
    if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
      this.showMessage('Loading...');
      messaging.peerSocket.send({});
    } else {
      this.showMessage('Not connected to peer.');
    }
  }
  
  handleMessage(evt) {
    if (evt.data['error']) {
      this.showMessage(evt.data['error']);
    } else {
      this.showResults(evt.data['data']);
    }
  }
}

const overview = new OverviewPage();

messaging.peerSocket.addEventListener("open", () => overview.refresh());
messaging.peerSocket.addEventListener("message", (evt) => {
  if (evt.data['service']) {
    console.log(evt.data);
  } else {
    overview.handleMessage(evt);
  }
});

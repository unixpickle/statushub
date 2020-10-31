import document from "document";
import * as messaging from "messaging";

class OverviewPage {
  constructor() {
    this.resultSections = document.getElementsByClassName('result-section');
    this.resultBoxes = document.getElementsByClassName('result-box');
    this.resultBoxTitles = document.getElementsByClassName('result-box-title');
    this.refreshButton = document.getElementById('refresh-button');

    this.refreshButton.addEventListener('click', () => this.refresh());
    this.resultSections.forEach((row, i) => {
      const title = row.getElementsByClassName('result-box-title')[0];
      const handler = () => {
        if (title.text) {
          showServiceLog(title.text);
        }
      };
      this.resultBoxes[i].addEventListener('click', handler);
      this.resultBoxTitles[i].addEventListener('click', handler);
    });

    this.showMessage('Not connected to peer.');
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

class ServiceLogPage {
  constructor(service) {
    this.service = service;
    this.container = document.getElementById('servicelog-page');
    this.resultBoxes = this.container.getElementsByClassName('result-box');
    this.refreshButton = this.container.getElementById('refresh-button');
    this.refreshButton.addEventListener('click', () => this.refresh());
  }

  clearResults() {
    this.resultBoxes.forEach((x) => x.text = '');
  }

  showMessage(msg) {
    this.clearResults();
    this.resultBoxes[0].text = msg;
  }

  refresh() {
    if (messaging.peerSocket.readyState === messaging.peerSocket.OPEN) {
      this.showMessage('Loading...');
      messaging.peerSocket.send({service: this.service});
    } else {
      this.showMessage('Not connected to peer.');
    }
  }

  showResults(logEntries) {
    this.clearResults();
    logEntries.forEach((x, i) => {
      if (i < this.resultBoxes.length) {
        this.resultBoxes[i].text = x;
      }
    });
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
let serviceLog = null;

function showServiceLog(service) {
  document.location.assign('./resources/servicelog.view').then(() => {
    serviceLog = new ServiceLogPage(service);
    serviceLog.refresh();
  });
}

messaging.peerSocket.addEventListener("open", () => overview.refresh());
messaging.peerSocket.addEventListener("message", (evt) => {
  if (evt.data['service']) {
    if (serviceLog && serviceLog.service === evt.data['service']) {
      serviceLog.handleMessage(evt);
    }
  } else {
    overview.handleMessage(evt);
  }
});

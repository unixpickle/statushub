import document from "document";
import * as messaging from "../common/bulk_messaging";

class ResultRow {
  constructor(elements) {
    this.elements = elements;
  }

  show(result) {
    if ('string' === typeof result) {
      this.elements[this.elements.length - 1].text = result;
    } else {
      result.forEach((x, i) => {
        this.elements[i].text = x;
      });
    }
  }

  clear() {
    this.elements.forEach((x) => x.text = '');
  }

  static findTitledRows() {
    const titles = document.getElementsByClassName('result-box-title');
    const values = document.getElementsByClassName('result-box');
    return titles.map((title, i) => new ResultRow([title, values[i]]));
  }

  static findTextRows() {
    const values = document.getElementsByClassName('result-box');
    return values.map((value) => new ResultRow([value]));
  }
}

class ListPage {
  constructor(resultRows) {
    this.resultRows = resultRows;
    this.refreshButton = document.getElementById('refresh-button');
    this.refreshButton.addEventListener('click', () => this.refresh());
    this.showMessage('Not connected to peer.');
  }

  clear() {
    this.resultRows.forEach((x) => x.clear());
  }

  show(results) {
    this.clear();
    results.forEach((x, i) => {
      if (i < this.resultRows.length) {
        this.resultRows[i].show(x);
      }
    });
  }

  showMessage(msg) {
    this.clear();
    this.resultRows[0].show(msg);
  }

  refresh() {
    if (messaging.peerSocket.isOpen()) {
      this.showMessage('Loading...');
      this._refresh();
    } else {
      this.showMessage('Not connected to peer.');
    }
  }

  handleMessage(evt) {
    if (evt.data['error']) {
      this.showMessage(evt.data['error']);
    } else {
      this.show(evt.data['data']);
    }
  }

  _refresh() {
    // Override this to send a request to the peer.
  }
}

class OverviewPage extends ListPage {
  constructor() {
    super(ResultRow.findTitledRows());

    this.resultRows.forEach((row) => {
      const title = row.elements[0];
      const handler = () => {
        if (title.text) {
          showServiceLog(title.text);
        }
      };
      row.elements.forEach((x) => x.addEventListener('click', handler));
    });
  }

  _refresh() {
    messaging.peerSocket.send({});
  }
}

class ServiceLogPage extends ListPage {
  constructor(service) {
    super(ResultRow.findTextRows());
    this.service = service;
  }

  _refresh() {
    messaging.peerSocket.send({service: this.service});
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

messaging.peerSocket.onOpen = () => overview.refresh();
messaging.peerSocket.onMessage = (evt) => {
  if (evt.data['service']) {
    if (serviceLog && serviceLog.service === evt.data['service']) {
      serviceLog.handleMessage(evt);
    }
  } else {
    overview.handleMessage(evt);
  }
};

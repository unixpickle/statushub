class Client {
  constructor() {
    this.close();
    this._stream = null;
  }

  fetchOverview() {
    callAPI('overview', {}, (e, d) => this.onOverview(e, d));
  }

  fetchServiceLog(name) {
    callAPI('serviceLog', { service: name }, (e, d) => this.onServiceLog(e, d));
  }

  fetchMediaOverview() {
    callAPI('mediaOverview', {}, (e, d) => this.onMediaOverview(e, d));
  }

  fetchMediaLog(name) {
    callAPI('mediaLog', { folder: name }, (e, d) => this.onMediaLog(e, d));
  }

  startStreaming(name) {
    if (this._stream) {
      this._stream.stop();
      this._stream = null;
    }
    this._stream = new ServiceStream(name);
    this._stream.onchange = (log) => this.onMediaLogStream(name, log);
    this._stream.onerror = (err) => {
      this._stream = null;
      this.onMediaLogStreamError(name, err);
    }
    this._stream.start();
  }

  stopStreaming() {
    if (this._stream) {
      this._stream.stop();
    }
    this._stream = null;
  }

  close() {
    this.onOverview = () => null;
    this.onServiceLog = () => null;
    this.onMediaOverview = () => null;
    this.onMediaLog = () => null;
    this.onMediaLogStream = () => null;
    this.onMediaLogStreamError = () => null;
  }
}

function callAPI(name, params, cb) {
  const req = new XMLHttpRequest();
  req.open('POST', '/api/' + name, true);
  req.setRequestHeader('Content-Type', 'application/json');
  req.onreadystatechange = () => {
    if (req.readyState === 4) {
      let obj;
      try {
        obj = JSON.parse(req.responseText);
      } catch (e) {
        cb('invalid JSON in response', null);
        return;
      }
      if (obj.error) {
        cb(obj.error, null);
      } else {
        cb(null, obj.data);
      }
    }
  };
  req.send(JSON.stringify(params));
  return req;
}


class ServiceStream {
  constructor(serviceName) {
    this.onchange = (_) => null;
    this.onerror = (_) => null;
    this.serviceName = serviceName;
    this._lastLog = null;
    this._socket = null;

    this._refreshWaiting = false;
    this._refreshNeeded = false;
    this._pendingEvents = [];
  }

  start() {
    const socket = new WebSocket(
      (location.protocol == 'https' ? 'wss' : 'ws') +
      '://' +
      location.host +
      '/api/serviceStream?service=' +
      encodeURIComponent(this.serviceName)
    )

    socket.addEventListener('open', () => {
      if (socket !== this._socket) {
        return;
      }
      this._refreshLog();
    });

    let firstMsg = true;
    socket.addEventListener('message', (event) => {
      if (socket !== this._socket) {
        return;
      }
      const msg = JSON.parse(event.data);
      if (firstMsg) {
        this._pendingEvents.push(msg);
        this._refreshLog();
        firstMsg = false;
      } else if (this._refreshWaiting) {
        this._pendingEvents.push(msg);
      } else {
        this._handleEvent(msg);
      }
    });

    this._socket = socket;
  }

  stop() {
    if (this._socket !== null) {
      this._socket.close();
      this._socket = null;
      this._refreshNeeded = false;
      this._pendingEvents = [];
    }
  }

  log() {
    return this._lastLog;
  }

  isRunning() {
    return this._socket !== null;
  }

  _refreshLog() {
    if (this._refreshWaiting) {
      this._refreshNeeded = true;
      return;
    }
    this._refreshWaiting = true;
    this._refreshNeeded = false;
    callAPI('serviceLog', { service: this.serviceName }, (e, d) => {
      this._refreshWaiting = false;
      if (!this.isRunning()) {
        this._refreshNeeded = false;
      } else if (e) {
        this._refreshNeeded = false;
        this.stop();
        this.onerror(e);
      } else {
        if (this._refreshNeeded) {
          // This refresh might contain stale data.
          this._refreshLog();
        } else {
          this._handleRefresh(d);
        }
        while (this._pendingEvents.length) {
          this._handleEvent(this._pendingEvents.shift());
        }
      }
    });
  }

  _handleRefresh(newLog) {
    if (this._lastLog === null) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
      return;
    }
    const idMap = {};
    this._lastLog.forEach((x) => {
      idMap[x.id] = true;
    });
    if (newLog.some((x) => !idMap[x.id])) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
    }
  }

  _handleEvent(data) {
    if (!this._lastLog.some((x) => x.id == data.id)) {
      this._lastLog.unshift(data);
      this.onchange(this._lastLog);
    }
  }
}


function mediaItemURL(id) {
  return '/api/mediaView?id=' + id;
}

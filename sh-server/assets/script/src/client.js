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


// Track the log messages in a service.
//
// After instantiating, set the onchange() and onerror() event handlers.
// The stream will continue to pass updated logs no onchange() until stop() is
// called or an error is encountered (passed through the onerror() event).
//
// When the stream is initiated, it will fetch the full service log before
// listening for events to make sure an accurate, full log is maintained.
// After this list is fetched, onchange() will be called.
class ServiceStream {
  constructor(serviceName) {
    this.onchange = (_) => null;
    this.onerror = (_) => null;
    this.serviceName = serviceName;
    this._lastLog = null;

    // This value is true while we are waiting for the full service
    // log request to complete.
    this._refreshWaiting = false;

    // This value is true if a full service log request was in progress
    // but we made another one (e.g. because the socket got a message).
    this._refreshNeeded = false;

    // Events are queued up here until the full service log request is
    // completed.
    this._pendingEvents = [];

    const socket = new WebSocket(
      (location.protocol == 'https:' ? 'wss' : 'ws') +
      '://' +
      location.host +
      '/api/serviceStream?service=' +
      encodeURIComponent(this.serviceName)
    )

    socket.addEventListener('open', () => {
      if (this.isRunning()) {
        this._refreshLog();
      }
    });

    let firstMsg = true;
    socket.addEventListener('message', (event) => {
      if (!this.isRunning()) {
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
        this._handleEvents([msg]);
      }
    });

    socket.addEventListener('close', () => {
      if (!this.isRunning()) {
        return;
      }
      this.stop();
      this.onerror('socket closed');
    });

    this._socket = socket;
  }

  isRunning() {
    return this._socket !== null;
  }

  stop() {
    if (this.isRunning()) {
      this._socket.close();
      this._socket = null;
    }
  }

  log() {
    return this._lastLog;
  }

  _refreshLog() {
    if (this._refreshWaiting) {
      this._refreshNeeded = true;
      return;
    }
    this._refreshWaiting = true;
    this._refreshNeeded = false;
    callAPI('serviceLog', { service: this.serviceName }, (e, d) => {
      if (!this.isRunning()) {
        return;
      } else if (e) {
        this.stop();
        this.onerror(e);
        return;
      }
      this._refreshWaiting = false;
      if (this._refreshNeeded) {
        // This refresh might contain stale data.
        this._refreshLog();
        return;
      }
      this._handleRefresh(d);
      const events = this._pendingEvents;
      this._pendingEvents = [];
      this._handleEvents(events);
    });
  }

  _handleRefresh(newLog) {
    if (this._lastLog === null) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
      return;
    }
    const idMap = {};
    this._lastLog.forEach((x) => idMap[x.id] = true);
    if (newLog.some((x) => !idMap[x.id])) {
      this._lastLog = newLog;
      this.onchange(this._lastLog);
    }
  }

  _handleEvents(datas) {
    let changed = false;
    const seenIDs = {};
    this._lastLog.forEach((x) => seenIDs[x.id] = true);
    datas.forEach((data) => {
      if (seenIDs[data.id]) {
        return;
      }
      seenIDs[data.id] = true;
      changed = true;
      let limit = 0;
      if (data.hasOwnProperty('limit')) {
        limit = data['limit'];
        delete data['limit'];
      }
      this._lastLog.unshift(data);
      while (limit && this._lastLog.length > limit) {
        this._lastLog.pop();
      }
    });
    if (changed) {
      this.onchange(this._lastLog);
    }
  }
}


function mediaItemURL(id) {
  return '/api/mediaView?id=' + id;
}

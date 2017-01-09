class Client {
  constructor() {
    this.close();
  }

  fetchOverview() {
    setTimeout(() => {
      this.onOverview(null, [
        {serviceName: 'FooService', id: 0, message: 'deleting /foo/bar'},
        {serviceName: 'NetService', id: 1, message: 'Current cost: 0.05'},
        {serviceName: 'NetService1', id: 2, message: 'Current cost: 0.92'}
      ]);
    }, 1000);
  }

  fetchServiceLog(name) {
    setTimeout(() => {
      this.onServiceLog(null, [
        {serviceName: name, id: 4, message: 'This is a log message.'},
        {serviceName: name, id: 5, message: 'The quick brown fox.'},
      ]);
    }, 1000);
  }

  fetchSettings() {
    setTimeout(() => {
      this.onSettings(null, {
        maxLog: 1000
      });
    });
  }

  fetchFullLog() {
    setTimeout(() => {
      this.onFullLog('network failure', null);
    }, 1000);
  }

  close() {
    this.onOverview = function() {};
    this.onServiceLog = function() {};
    this.onFullLog = function() {};
    this.onSettings = function() {};
  }
}

function callAPI(name, params, cb) {
  let canceled = false;
  const req = new XMLHttpRequest();
  req.open('POST', '/api/'+name, true);
  req.setRequestHeader('Content-Type', 'application/json');
  req.onreadystatechange = () => {
    if (req.readyState === 4) {
      try {
        const obj = JSON.parse(req.responseText);
        if (obj.error) {
          cb(obj.error, null);
        } else {
          cb(null, obj.data);
        }
      } catch (e) {
        cb('invalid JSON in response', null);
      }
    }
  };
  req.send(JSON.stringify(params));
  return req;
}

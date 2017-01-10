class Client {
  constructor() {
    this.close();
  }

  fetchOverview() {
    callAPI('overview', {}, (e, d) => this.onOverview(e, d));
  }

  fetchServiceLog(name) {
    callAPI('serviceLog', {service: name}, (e, d) => this.onServiceLog(e, d));
  }

  fetchFullLog() {
    callAPI('fullLog', {}, (e, d) => this.onFullLog(e, d));
  }

  close() {
    this.onOverview = function() {};
    this.onServiceLog = function() {};
    this.onFullLog = function() {};
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

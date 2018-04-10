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

  fetchMediaOverview() {
    callAPI('mediaOverview', {}, (e, d) => this.onMediaOverview(e, d));
  }

  fetchMediaLog(name) {
    callAPI('mediaLog', {folder: name}, (e, d) => this.onMediaLog(e, d));
  }

  close() {
    this.onOverview = function() {};
    this.onServiceLog = function() {};
    this.onMediaOverview = function() {};
    this.onMediaLog = function() {};
  }
}

function callAPI(name, params, cb) {
  let canceled = false;
  const req = new XMLHttpRequest();
  req.open('POST', '/api/'+name, true);
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

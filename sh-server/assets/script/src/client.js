class Client {
  constructor() {
    this.onOverview = function() {};
    this.onServiceLog = function() {};
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
    // TODO: this.
  }

  fetchFullLog() {
    // TODO: this.
  }

  close() {
    this.onOverview = function() {};
  }
}

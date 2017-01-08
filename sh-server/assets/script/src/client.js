class Client {
  constructor() {
    this.onOverview = function() {};

    setTimeout(() => {
      this.onOverview(null, [
        {serviceName: 'FooService', id: 0, message: 'deleting /foo/bar'},
        {serviceName: 'NetService', id: 1, message: 'Current cost: 0.05'},
        {serviceName: 'NetService1', id: 2, message: 'Current cost: 0.92'}
      ]);
    }, 1000);
  }

  close() {
    this.onOverview = function() {};
  }
}

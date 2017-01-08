class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: { error: null, entries: null },
      fullLog: { error: null, entries: null },
      serviceLog: { error: null, entries: null },
      settings: { error: null, entries: null },
      serviceLogReq: ''
    };
    if (!history.state) {
      this.replaceHistory();
    } else {
      this.state.page = history.state.page;
      this.state.serviceLogReq = history.state.serviceLogReq;
    }
    this._client = null;
  }

  componentDidMount() {
    this._client = new Client();
    this._client.onOverview = this.gotSceneData.bind(this, 'overview');
    this._client.onServiceLog = this.gotSceneData.bind(this, 'serviceLog');

    window.onpopstate = e => {
      this.setState(e.state, () => this.fetchPageData());
    };

    this.fetchPageData();
  }

  componentWillUnmount() {
    this._client.close();
  }

  render() {
    return React.createElement(
      'div',
      null,
      React.createElement(NavBar, { page: this.state.page }),
      this.pageContent()
    );
  }

  pageContent() {
    switch (this.state.page) {
      case 'overview':
        return React.createElement(LogScene, { info: this.state.overview,
          onClick: e => this.showServiceLog(e) });
      case 'fullLog':
        // TODO: this.
        break;
      case 'serviceLog':
        return React.createElement(LogScene, { info: this.state.serviceLog });
      case 'settings':
        // TODO: this.
        break;
    }
    throw new Error('unsupported page: ' + this.state.page);
  }

  fetchPageData() {
    if (this.state[this.state.page].entries) {
      return;
    }
    switch (this.state.page) {
      case 'overview':
        this._client.fetchOverview();
        break;
      case 'fullLog':
        this._client.fetchFullLog();
        break;
      case 'settings':
        this._client.fetchSettings();
        break;
      case 'serviceLog':
        this._client.fetchServiceLog(this.state.serviceLogReq);
        break;
    }
  }

  gotSceneData(name, err, data) {
    const obj = {};
    obj[name] = { error: err, entries: data };
    if (name === this.state.page) {
      this.setState(obj, () => this.replaceHistory());
    } else {
      this.setState(obj);
    }
  }

  showServiceLog(info) {
    this.setState({
      page: 'serviceLog',
      serviceLog: { error: null, entries: null },
      serviceLogReq: info.serviceName
    }, () => {
      this.pushHistory();
      this.fetchPageData();
    });
  }

  pushHistory() {
    history.pushState(this.historyState(), window.title, this.pageHash());
  }

  replaceHistory() {
    history.replaceState(this.historyState(), window.title, this.pageHash());
  }

  pageHash() {
    if (this.state.page === 'overview') {
      return '';
    }
    return '#' + this.state.page;
  }

  historyState() {
    return {
      page: this.state.page,
      serviceLogReq: this.state.serviceLogReq
    };
  }
}

window.addEventListener('load', function () {
  ReactDOM.render(React.createElement(Root, null), document.getElementById('root'));
});
class Client {
  constructor() {
    this.onOverview = function () {};
    this.onServiceLog = function () {};
  }

  fetchOverview() {
    setTimeout(() => {
      this.onOverview(null, [{ serviceName: 'FooService', id: 0, message: 'deleting /foo/bar' }, { serviceName: 'NetService', id: 1, message: 'Current cost: 0.05' }, { serviceName: 'NetService1', id: 2, message: 'Current cost: 0.92' }]);
    }, 1000);
  }

  fetchServiceLog(name) {
    setTimeout(() => {
      this.onServiceLog(null, [{ serviceName: name, id: 4, message: 'This is a log message.' }, { serviceName: name, id: 5, message: 'The quick brown fox.' }]);
    }, 1000);
  }

  fetchSettings() {
    // TODO: this.
  }

  fetchFullLog() {
    // TODO: this.
  }

  close() {
    this.onOverview = function () {};
  }
}
function Loader(props) {
  return React.createElement(
    'div',
    { className: 'loader' },
    'Loading'
  );
}
function LogScene(props) {
  const info = props.info;
  if (info.entries) {
    return React.createElement(LogPane, { items: info.entries, onClick: props.onClick });
  } else if (info.error) {
    return React.createElement(
      'div',
      { className: 'log-scene-error' },
      React.createElement(
        'label',
        null,
        'Load failed: ',
        info.error
      )
    );
  } else {
    return React.createElement(
      'div',
      { className: 'log-scene-loading' },
      React.createElement(Loader, null)
    );
  }
}

function LogPane(props) {
  const items = props.items.map(x => {
    return React.createElement(LogItem, { info: x, key: x.id, onClick: props.onClick });
  });
  return React.createElement(
    'ul',
    { className: 'log-pane' },
    items
  );
}

function LogItem(props) {
  const inf = props.info;
  const clickHandler = () => props.onClick(inf);
  return React.createElement(
    'li',
    { className: props.onClick ? 'clickable' : '', onClick: clickHandler },
    React.createElement(
      'label',
      { className: 'service-name' },
      inf.serviceName
    ),
    React.createElement(
      'label',
      { className: 'message' },
      inf.message
    )
  );
}
function NavBar(props) {
  const page = props.page;
  return React.createElement(
    'nav',
    null,
    React.createElement(VoidLink, { onClick: props.onOverview, name: 'Overview', cur: page === 'overview' }),
    React.createElement(VoidLink, { onClick: props.onFullLog, name: 'Full Log', cur: page === 'fullLog' }),
    React.createElement(VoidLink, { onClick: props.onSettings, name: 'Settings', cur: page === 'settings' })
  );
}

function VoidLink(props) {
  if (!props.cur) {
    return React.createElement(
      'a',
      { href: 'javascript:void(0)', onClick: props.onClick },
      props.name
    );
  } else {
    return React.createElement(
      'a',
      { href: 'javascript:void(0)', className: 'cur',
        onClick: props.onClick },
      props.name
    );
  }
}

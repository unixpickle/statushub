class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: { error: null, services: null },
      fullLog: { error: null, entries: null },
      serviceLog: { error: null, entries: null },
      settings: { error: null, settings: null }
    };
    this._client = null;
  }

  componentDidMount() {
    this._client = new Client();
    this._client.onOverview = (err, obj) => {
      this.setState({
        overview: {
          error: err,
          services: obj
        }
      });
    };
  }

  componentWillUnmount() {
    this._client.close();
  }

  showServiceLog(id) {
    console.log('got ID: ' + id);
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
        return React.createElement(Overview, { info: this.state.overview, onClick: id => this.showServiceLog(id) });
      case 'fullLog':
        // TODO: this.
        break;
      case 'serviceLog':
        // TODO: this.
        break;
      case 'settings':
        // TODO: this.
        break;
    }
    throw new Error('unsupported page: ' + this.state.page);
  }
}

window.addEventListener('load', function () {
  ReactDOM.render(React.createElement(Root, null), document.getElementById('root'));
});
class Client {
  constructor() {
    this.onOverview = function () {};

    setTimeout(() => {
      this.onOverview(null, [{ serviceName: 'FooService', id: 0, message: 'deleting /foo/bar' }, { serviceName: 'NetService', id: 1, message: 'Current cost: 0.05' }, { serviceName: 'NetService1', id: 2, message: 'Current cost: 0.92' }]);
    }, 1000);
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
  const clickHandler = () => props.onClick(inf.id);
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
function Overview(props) {
  const info = props.info;
  if (info.services) {
    return React.createElement(LogPane, { items: info.services, onClick: props.onClick });
  } else if (info.error) {
    return React.createElement(
      'div',
      { className: 'overview-error' },
      React.createElement(
        'label',
        null,
        info.error
      )
    );
  } else {
    return React.createElement(
      'div',
      { className: 'overview-loading' },
      React.createElement(Loader, null)
    );
  }
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

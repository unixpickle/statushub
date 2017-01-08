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
        return React.createElement(Overview, { info: this.state.overview });
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
function Overview(props) {
  const info = props.info;
  if (info.services) {
    const items = info.services.map(x => {
      return React.createElement(OverviewItem, { info: x, key: x.info.serviceName });
    });
    return React.createElement(
      'ul',
      { className: 'overview-list' },
      items
    );
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

function OverviewItem(props) {
  return React.createElement(
    'li',
    { 'class': 'overview-item' },
    React.createElement(
      'label',
      { 'class': 'service-name' },
      'props.info.serviceName'
    ),
    React.createElement(
      'label',
      { 'class': 'message' },
      'props.info.message'
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

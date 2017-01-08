class Root extends React.Component {
  constructor() {
    super();
    this.state = {
      page: 'overview',
      overview: {error: null, services: null},
      fullLog: {error: null, entries: null},
      serviceLog: {error: null, entries: null},
      settings: {error: null, settings: null}
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
    return (
      <div>
        <NavBar page={this.state.page} />
        {this.pageContent()}
      </div>
    );
  }

  pageContent() {
    switch (this.state.page) {
    case 'overview':
      return <Overview info={this.state.overview} onClick={(id) => this.showServiceLog(id)} />;
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

window.addEventListener('load', function() {
  ReactDOM.render(<Root />, document.getElementById('root'));
});

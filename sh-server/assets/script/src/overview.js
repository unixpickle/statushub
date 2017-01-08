function Overview(props) {
  const info = props.info;
  if (info.services) {
    return <LogPane items={info.services} onClick={props.onClick} />;
  } else if (info.error) {
    return (
      <div className="overview-error">
        <label>{info.error}</label>
      </div>
    );
  } else {
    return (
      <div className="overview-loading">
        <Loader />
      </div>
    );
  }
}

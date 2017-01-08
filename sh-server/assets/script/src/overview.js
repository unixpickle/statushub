function Overview(props) {
  const info = props.info;
  if (info.services) {
    const items = info.services.map((x) => {
      return <OverviewItem info={x} key={x.info.serviceName} />;
    });
    return <ul className="overview-list">{items}</ul>;
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

function OverviewItem(props) {
  return (
    <li class="overview-item">
      <label class="service-name">props.info.serviceName</label>
      <label class="message">props.info.message</label>
    </li>
  );
}

function LogPane(props) {
  const items = props.items.map((x) => {
    return <LogItem info={x} key={x.id} onClick={props.onClick} />;
  });
  return <ul className="log-pane">{items}</ul>;
}

function LogItem(props) {
  const inf = props.info;
  const clickHandler = () => props.onClick(inf.id);
  return (
    <li className={props.onClick ? 'clickable' : ''} onClick={clickHandler}>
      <label className="service-name">{inf.serviceName}</label>
      <label className="message">{inf.message}</label>
    </li>
  );
}

function NavBar(props) {
  const page = props.page;
  return (
    <nav>
      <VoidLink onClick={props.onOverview} name="Overview" cur={page === 'overview'}/>
      <VoidLink onClick={props.onFullLog} name="Full Log" cur={page === 'fullLog'}/>
      <VoidLink onClick={props.onSettings} name="Settings" cur={page === 'settings'} />
    </nav>
  );
}

function VoidLink(props) {
  if (!props.cur) {
    return <a href="javascript:void(0)" onClick={props.onClick}>{props.name}</a>;
  } else {
    return <a href="javascript:void(0)" className="cur"
      onClick={props.onClick}>{props.name}</a>;
  }
}

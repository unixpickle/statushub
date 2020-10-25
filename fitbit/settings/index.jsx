function mySettings(props) {
  return (
    <Page>
      <Section
        title={<Text bold align="center">Login Settings</Text>}>
        <TextInput
          label="Base URL"
          settingsKey="baseurl"
          placeholder="https://statushub.mywebsite.com" />
        <TextInput
          label="Password"
          settingsKey="password"
          placeholder="My Password"
          type="password" />
      </Section>
    </Page>
  );
}

registerSettingsPage(mySettings);

import React, { useEffect, useState } from 'react';
import { Divider, Form, Grid, Header } from 'semantic-ui-react';
import { API, showError } from '../helpers';

const OtherSetting = () => {
  let [inputs, setInputs] = useState({
    Footer: '',
    Notice: '',
    About: '',
    HomePageLink: '',
  });
  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key in inputs) {
          newInputs[item.key] = item.value;
        }
      });
      setInputs(newInputs);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submitNotice = async () => {
    await updateOption('Notice', inputs.Notice);
  };

  const submitFooter = async () => {
    await updateOption('Footer', inputs.Footer);
  };

  const submitHomePageLink = async () => {
    await updateOption('HomePageLink', inputs.HomePageLink);
  };

  const submitAbout = async () => {
    await updateOption('About', inputs.About);
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading}>
          <Header as='h3'>通用设置</Header>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='公告'
              placeholder='在此输入新的公告内容'
              value={inputs.Notice}
              name='Notice'
              onChange={handleInputChange}
              style={{ minHeight: 150, fontFamily: 'JetBrains Mono, Consolas' }}
            />
          </Form.Group>
          <Form.Button onClick={submitNotice}>保存公告</Form.Button>
          <Divider />
          <Header as='h3'>个性化设置</Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='首页链接'
              placeholder='在此输入首页链接，设置后将通过 iframe 方式嵌入该网页'
              value={inputs.HomePageLink}
              name='HomePageLink'
              onChange={handleInputChange}
              type='url'
            />
          </Form.Group>
          <Form.Button onClick={submitHomePageLink}>设置首页链接</Form.Button>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='关于'
              placeholder='在此输入新的关于内容，支持 Markdown & HTML 代码'
              value={inputs.About}
              name='About'
              onChange={handleInputChange}
              style={{ minHeight: 150, fontFamily: 'JetBrains Mono, Consolas' }}
            />
          </Form.Group>
          <Form.Button onClick={submitAbout}>保存关于</Form.Button>
          <Form.Group widths='equal'>
            <Form.Input
              label='页脚'
              placeholder='在此输入页脚内容，留空则不显示，支持 HTML 代码'
              value={inputs.Footer}
              name='Footer'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitFooter}>设置页脚</Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default OtherSetting;

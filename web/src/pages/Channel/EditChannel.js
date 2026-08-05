import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Message, Segment } from 'semantic-ui-react';
import { useParams, Link } from 'react-router-dom';
import { API, generateToken, showError, showSuccess } from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants';

const EditChannel = () => {
  const params = useParams();
  const channelId = params.id;
  const isEditing = channelId !== undefined;
  const [loading, setLoading] = useState(isEditing);
  const originInputs = {
    type: 'none',
    name: '',
    description: '',
    secret: '',
    app_id: '',
    account_id: '',
    url: '',
    other: '',
    corp_id: '', // only for corp_app
    agent_id: '', // only for corp_app
    token: '',
    bark_group: '',
    bark_sound: '',
    bark_icon: '',
    bark_level: '',
    bark_is_archive: '',
    bark_url: '',
    ntfy_priority: '',
    ntfy_icon: '',
    ntfy_username: '',
    ntfy_password: '',
    gotify_priority: '',
    pushme_date: '',
    pushme_type: '',
    corp_client_type: 'app',
    corp_proxy: '',
    custom_headers: '',
  };

  const [inputs, setInputs] = useState(originInputs);
  const { type, name, description, secret, app_id, account_id, url, other } =
    inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadChannel = async () => {
    let res = await API.get(`/api/channel/${channelId}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.type === 'corp_app') {
        const [corp_id, agent_id] = data.app_id.split('|');
        data.corp_id = corp_id;
        data.agent_id = agent_id;
        data.corp_client_type = 'app';
        data.corp_proxy = '';
        if (data.other) {
          try {
            const opts = JSON.parse(data.other);
            data.corp_client_type = opts.client_type || 'app';
            data.corp_proxy = opts.proxy || '';
          } catch (e) {
            // keep defaults
          }
        }
      }
      if (data.type === 'bark' && data.other) {
        try {
          const opts = JSON.parse(data.other);
          data.bark_group = opts.group || '';
          data.bark_sound = opts.sound || '';
          data.bark_icon = opts.icon || '';
          data.bark_level = opts.level || '';
          data.bark_is_archive = opts.is_archive || '';
          data.bark_url = opts.url || '';
        } catch (e) {
          // keep raw other for unexpected legacy values
        }
      }
      if (data.type === 'ntfy' && data.other) {
        try {
          const opts = JSON.parse(data.other);
          data.ntfy_priority = opts.priority || '';
          data.ntfy_icon = opts.icon || '';
          data.ntfy_username = opts.username || '';
          data.ntfy_password = opts.password || '';
        } catch (e) {
          // ignore invalid json
        }
      }
      if (data.type === 'gotify' && data.other) {
        try {
          const opts = JSON.parse(data.other);
          data.gotify_priority = opts.priority || '';
        } catch (e) {
          // ignore invalid json
        }
      }
      if (data.type === 'pushme' && data.other) {
        try {
          const opts = JSON.parse(data.other);
          data.pushme_date = opts.date || '';
          data.pushme_type = opts.type || '';
        } catch (e) {
          // ignore invalid json
        }
      }
      if (data.type === 'custom' && data.other) {
        data.custom_headers = '';
        try {
          const opts = JSON.parse(data.other);
          if (
            opts &&
            typeof opts === 'object' &&
            Object.prototype.hasOwnProperty.call(opts, 'headers') &&
            Object.prototype.hasOwnProperty.call(opts, 'body') &&
            typeof opts.body === 'string'
          ) {
            data.custom_headers =
              opts.headers && Object.keys(opts.headers).length > 0
                ? JSON.stringify(opts.headers, null, 2)
                : '';
            data.other = opts.body;
          }
        } catch (e) {
          // keep raw body template
        }
      }
      setInputs({ ...originInputs, ...data });
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    if (isEditing) {
      loadChannel().then();
    }
  }, []);

  const submit = async () => {
    if (!name) return;
    let res = undefined;
    let localInputs = { ...inputs };
    switch (inputs.type) {
      case 'corp_app':
        localInputs.app_id = `${inputs.corp_id}|${inputs.agent_id}`;
        {
          const corpOpts = {
            client_type: localInputs.corp_client_type || 'app',
          };
          if (localInputs.corp_proxy)
            corpOpts.proxy = localInputs.corp_proxy.trim();
          localInputs.other = JSON.stringify(corpOpts);
          if (localInputs.url) {
            localInputs.url = localInputs.url.replace(/\/+$/, '');
          }
        }
        break;
      case 'bark':
        if (localInputs.url === '') {
          localInputs.url = 'https://api.day.app';
        }
        {
          const barkOpts = {};
          if (localInputs.bark_group) barkOpts.group = localInputs.bark_group.trim();
          if (localInputs.bark_sound) barkOpts.sound = localInputs.bark_sound.trim();
          if (localInputs.bark_icon) barkOpts.icon = localInputs.bark_icon.trim();
          if (localInputs.bark_level) barkOpts.level = localInputs.bark_level.trim();
          if (localInputs.bark_is_archive)
            barkOpts.is_archive = localInputs.bark_is_archive.trim();
          if (localInputs.bark_url) barkOpts.url = localInputs.bark_url.trim();
          localInputs.other =
            Object.keys(barkOpts).length > 0 ? JSON.stringify(barkOpts) : '';
        }
        break;
      case 'ntfy':
        if (localInputs.url === '') {
          localInputs.url = 'https://ntfy.sh';
        } else if (localInputs.url.endsWith('/')) {
          localInputs.url = localInputs.url.slice(0, -1);
        }
        if (!localInputs.account_id) {
          showError('请填写 ntfy topic！');
          return;
        }
        {
          const ntfyOpts = {};
          if (localInputs.ntfy_priority)
            ntfyOpts.priority = localInputs.ntfy_priority.trim();
          if (localInputs.ntfy_icon) ntfyOpts.icon = localInputs.ntfy_icon.trim();
          if (localInputs.ntfy_username)
            ntfyOpts.username = localInputs.ntfy_username.trim();
          if (localInputs.ntfy_password)
            ntfyOpts.password = localInputs.ntfy_password.trim();
          localInputs.other =
            Object.keys(ntfyOpts).length > 0 ? JSON.stringify(ntfyOpts) : '';
        }
        break;
      case 'gotify':
        if (localInputs.url === '') {
          showError('请填写 Gotify 服务器地址！');
          return;
        }
        if (localInputs.url.endsWith('/')) {
          localInputs.url = localInputs.url.slice(0, -1);
        }
        if (!localInputs.secret) {
          showError('请填写 Gotify 应用令牌！');
          return;
        }
        {
          const gotifyOpts = {};
          if (localInputs.gotify_priority)
            gotifyOpts.priority = localInputs.gotify_priority.trim();
          localInputs.other =
            Object.keys(gotifyOpts).length > 0
              ? JSON.stringify(gotifyOpts)
              : '';
        }
        break;
      case 'pushme':
        if (localInputs.url === '') {
          localInputs.url = 'https://push.i-i.me/';
        }
        if (!localInputs.secret) {
          showError('请填写 PushMe push_key！');
          return;
        }
        {
          const pushMeOpts = {};
          if (localInputs.pushme_date)
            pushMeOpts.date = localInputs.pushme_date.trim();
          if (localInputs.pushme_type)
            pushMeOpts.type = localInputs.pushme_type.trim();
          localInputs.other =
            Object.keys(pushMeOpts).length > 0
              ? JSON.stringify(pushMeOpts)
              : '';
        }
        break;
      case 'telegram':
        if (localInputs.url === '') {
          localInputs.url = 'https://api.telegram.org';
        } else if (localInputs.url.endsWith('/')) {
          localInputs.url = localInputs.url.slice(0, -1);
        }
        if (localInputs.other) {
          localInputs.other = localInputs.other.trim();
        }
        break;
      case 'discord':
        if (localInputs.other) {
          localInputs.other = localInputs.other.trim();
        }
        break;
      case 'one_bot':
        if (localInputs.url.endsWith('/')) {
          localInputs.url = localInputs.url.slice(0, -1);
        }
        break;
      case 'group':
        let channels = localInputs.app_id.split('|');
        let targets = localInputs.account_id.split('|');
        if (localInputs.account_id === '') {
          for (let i = 0; i < channels.length - 1; i++) {
            localInputs.account_id += '|';
          }
        } else if (channels.length !== targets.length) {
          showError(
            '群组通道的子通道数量与目标数量不匹配，对于不需要指定的目标请直接留空'
          );
          return;
        }
        break;
      case 'custom':
        // if (!localInputs.url.startsWith('https://')) {
        //   showError('自定义通道的 URL 必须以 https:// 开头！');
        //   return;
        // }
        try {
          JSON.parse(localInputs.other);
        } catch (e) {
          showError('请求体 JSON 格式错误：' + e.message);
          return;
        }
        {
          const headersText = (localInputs.custom_headers || '').trim();
          if (headersText) {
            let headersObj;
            try {
              headersObj = JSON.parse(headersText);
            } catch (e) {
              showError('请求头 JSON 格式错误：' + e.message);
              return;
            }
            if (
              !headersObj ||
              typeof headersObj !== 'object' ||
              Array.isArray(headersObj)
            ) {
              showError('请求头必须是 JSON 对象，例如 {"Authorization":"Bearer xxx"}');
              return;
            }
            localInputs.other = JSON.stringify({
              headers: headersObj,
              body: localInputs.other,
            });
          }
        }
        break;
    }
    if (isEditing) {
      res = await API.put(`/api/channel/`, {
        ...localInputs,
        id: parseInt(channelId),
      });
    } else {
      res = await API.post(`/api/channel`, localInputs);
    }
    const { success, message } = res.data;
    if (success) {
      if (isEditing) {
        showSuccess('通道信息更新成功！');
      } else {
        showSuccess('通道创建成功！');
        setInputs(originInputs);
      }
    } else {
      showError(message);
    }
  };

  const getTelegramChatId = async () => {
    if (inputs.secret === '') {
      showError('请先输入 Telegram 机器人令牌！');
      return;
    }
    try {
      let res = await API.post('/api/channel/telegram/chat_id', {
        secret: inputs.secret,
        url: inputs.url,
        other: inputs.other,
      });
      const { success, message, data } = res.data;
      if (success) {
        setInputs((inputs) => ({ ...inputs, account_id: data }));
        showSuccess('会话 ID 获取成功！');
      } else {
        showError(message);
      }
    } catch (e) {
      showError(`请求失败：${e.message}`);
    }
  };

  const renderChannelForm = () => {
    switch (type) {
      case 'email':
        return (
          <>
            <Message>
              邮件推送方式（email）需要设置邮箱，请前往个人设置页面绑定邮箱地址，之后系统将自动为你创建邮箱推送通道。
            </Message>
          </>
        );
      case 'test':
        return (
          <>
            <Message>
              通过微信测试号进行推送，点击前往配置：
              <a
                target='_blank'
                href='https://mp.weixin.qq.com/debug/cgi-bin/sandboxinfo?action=showinfo&t=sandbox/index'
              >
                微信公众平台接口测试帐号
              </a>
              。
              <br />
              需要新增测试模板，模板标题推荐填写为「消息推送」，模板内容填写为：
              <br />
              标题：{' {{'}title.DATA{'}}'}
              <br />
              描述：{' {{'}description.DATA{'}}'}
              <br />
              内容：{' {{'}content.DATA{'}}'}
              <br />
              用户 Open ID 支持单个、多个（用{' '}
              <code>|</code> 分隔）或 <code>@all</code>
              （推送给所有已关注用户）。推送参数 <code>to</code>{' '}
              同样支持上述写法。
            </Message>
            <Form.Group widths={3}>
              <Form.Input
                label='测试号 ID'
                name='app_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.app_id}
                placeholder='测试号信息 -> appID'
              />
              <Form.Input
                label='测试号密钥'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='测试号信息 -> appsecret'
              />
              <Form.Input
                label='测试模板 ID'
                name='other'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.other}
                placeholder='模板消息接口 -> 模板 ID'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.Input
                label='用户 Open ID'
                name='account_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='单个 OpenID，或多个用 | 分隔，或填 @all'
              />
            </Form.Group>
          </>
        );
      case 'corp_app':
        return (
          <>
            <Message>
              通过企业微信应用号进行推送，点击前往配置：
              <a
                target='_blank'
                href='https://work.weixin.qq.com/wework_admin/frame#apps'
              >
                企业微信应用管理
              </a>
              。
              <br />
              <br />
              注意，企业微信要求配置可信 IP。若服务在内网、出口 IP
              不便加白，可填写代理地址（仅该渠道走代理），或自定义 API
              地址指向已加白的反向代理 / 专有云入口。
            </Message>
            <Form.Group widths={3}>
              <Form.Input
                label='企业 ID'
                name='corp_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.corp_id}
                placeholder='我的企业 -> 企业信息 -> 企业 ID'
              />
              <Form.Input
                label='应用 AgentId'
                name='agent_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.agent_id}
                placeholder='应用管理 -> 自建 -> 创建应用 -> AgentId'
              />
              <Form.Input
                label='应用 Secret'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='应用管理 -> 自建 -> 创建应用 -> Secret'
              />
            </Form.Group>
            <Form.Group widths={3}>
              <Form.Input
                label='用户账号'
                name='account_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='通讯录 -> 点击姓名 -> 账号'
              />
              <Form.Select
                label='微信企业号客户端类型'
                name='corp_client_type'
                options={[
                  {
                    key: 'plugin',
                    text: '微信中的企业微信插件',
                    value: 'plugin',
                  },
                  { key: 'app', text: '企业微信 APP', value: 'app' },
                ]}
                value={inputs.corp_client_type}
                onChange={handleInputChange}
              />
              <Form.Input
                label='自定义 API 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='可选，默认 https://qyapi.weixin.qq.com'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.Input
                label='代理地址'
                name='corp_proxy'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.corp_proxy}
                placeholder='可选，如 http://127.0.0.1:7890 或 socks5://127.0.0.1:1080'
              />
            </Form.Group>
          </>
        );
      case 'corp':
        return (
          <>
            <Message>
              通过企业微信群机器人进行推送，配置流程：选择一个群聊 -&gt; 设置 -&gt;
              群机器人 -&gt; 添加 -&gt; 新建 -&gt; 输入名字，点击添加 -&gt; 点击复制 Webhook
              地址
            </Message>
            <Form.Group widths='equal'>
              <Form.Input
                label='Webhook 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写企业微信提供的 Webhook 地址'
              />
            </Form.Group>
          </>
        );
      case 'lark':
        return (
          <>
            <Message>
              通过飞书群机器人进行推送，飞书桌面客户端的配置流程：选择一个群聊
              -&gt; 设置 -&gt; 群机器人 -&gt; 添加机器人 -&gt; 自定义机器人 -&gt; 添加（
              <strong>注意选中「签名校验」</strong>）。具体参见：
              <a
                target='_blank'
                href='https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN'
              >
                飞书开放文档
              </a>
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='Webhook 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写飞书提供的 Webhook 地址'
              />
              <Form.Input
                label='签名校验密钥'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此填写飞书提供的签名校验密钥'
              />
            </Form.Group>
          </>
        );
      case 'ding':
        return (
          <>
            <Message>
              通过钉钉群机器人进行推送，钉钉桌面客户端的配置流程：选择一个群聊
              -&gt; 群设置 -&gt; 智能群助手 -&gt; 添加机器人（点击右侧齿轮图标） -&gt;
              自定义 -&gt; 添加（
              <strong>注意选中「加密」</strong>）。具体参见：
              <a
                target='_blank'
                href='https://open.dingtalk.com/document/robots/custom-robot-access'
              >
                钉钉开放文档
              </a>
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='Webhook 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写钉钉提供的 Webhook 地址'
              />
              <Form.Input
                label='签名校验密钥'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此填写钉钉提供的签名校验密钥'
              />
            </Form.Group>
          </>
        );
      case 'bark':
        return (
          <>
            <Message>
              通过 Bark 进行推送，下载 Bark 后按提示注册设备，之后会看到一个
              URL，例如 <code>https://api.day.app/wrsVSDRANDOM/Body Text</code>
              ，其中 <code>wrsVSDRANDOM</code> 就是你的推送 key。可选配置会随渠道保存，推送时自动带上。
              跳转 URL 仅在消息未指定自定义链接时生效（系统自动生成的详情页链接会被覆盖）。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='服务器地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写 Bark 服务器地址，不填则使用默认值'
              />
              <Form.Input
                label='推送 key'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此填写 Bark 推送 key'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='分组'
                name='bark_group'
                onChange={handleInputChange}
                value={inputs.bark_group}
                placeholder='推送分组，可选'
              />
              <Form.Input
                label='存档'
                name='bark_is_archive'
                onChange={handleInputChange}
                value={inputs.bark_is_archive}
                placeholder='1（存档）或 0（不存档），可选'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='推送图标'
                name='bark_icon'
                onChange={handleInputChange}
                value={inputs.bark_icon}
                placeholder='推送图标 URL，可选'
              />
              <Form.Input
                label='推送声音'
                name='bark_sound'
                onChange={handleInputChange}
                value={inputs.bark_sound}
                placeholder='推送铃声，可选'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='跳转 URL'
                name='bark_url'
                onChange={handleInputChange}
                value={inputs.bark_url}
                placeholder='点击推送跳转的 URL，可选'
              />
              <Form.Input
                label='推送时效'
                name='bark_level'
                onChange={handleInputChange}
                value={inputs.bark_level}
                placeholder='active / timeSensitive / passive，可选'
              />
            </Form.Group>
          </>
        );
      case 'ntfy':
        return (
          <>
            <Message>
              通过{' '}
              <a href='https://ntfy.sh' target='_blank' rel='noreferrer'>
                ntfy
              </a>{' '}
              推送（可自建）。Topic 必填；Token 与用户名/密码二选一即可。优先级为
              1-5（默认 3）。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='服务器地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='不填则使用 https://ntfy.sh'
              />
              <Form.Input
                label='Topic'
                name='account_id'
                onChange={handleInputChange}
                value={inputs.account_id}
                placeholder='在此填写订阅主题'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='Access Token'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='可选，Bearer Token'
              />
              <Form.Input
                label='优先级'
                name='ntfy_priority'
                onChange={handleInputChange}
                value={inputs.ntfy_priority}
                placeholder='1-5，可选，默认 3'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='用户名'
                name='ntfy_username'
                onChange={handleInputChange}
                value={inputs.ntfy_username}
                placeholder='可选，Basic Auth'
              />
              <Form.Input
                label='密码'
                name='ntfy_password'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.ntfy_password}
                placeholder='可选，Basic Auth'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.Input
                label='图标 URL'
                name='ntfy_icon'
                onChange={handleInputChange}
                value={inputs.ntfy_icon}
                placeholder='可选'
              />
            </Form.Group>
          </>
        );
      case 'gotify':
        return (
          <>
            <Message>
              通过{' '}
              <a href='https://gotify.net' target='_blank' rel='noreferrer'>
                Gotify
              </a>{' '}
              自托管推送。填写服务器地址与应用令牌（Apps → Create
              application）。优先级一般为 0-10，可选。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='服务器地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='例如 https://gotify.example.com'
              />
              <Form.Input
                label='应用令牌'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='Gotify application token'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.Input
                label='优先级'
                name='gotify_priority'
                onChange={handleInputChange}
                value={inputs.gotify_priority}
                placeholder='0-10，可选，默认 0'
              />
            </Form.Group>
          </>
        );
      case 'pushme':
        return (
          <>
            <Message>
              通过{' '}
              <a href='https://push.i-i.me/' target='_blank' rel='noreferrer'>
                PushMe
              </a>{' '}
              推送（可自建）。Push Key 必填；自建时填写自定义 API 地址。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='API 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='不填则使用 https://push.i-i.me/'
              />
              <Form.Input
                label='Push Key'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此填写 PushMe push_key'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='日期'
                name='pushme_date'
                onChange={handleInputChange}
                value={inputs.pushme_date}
                placeholder='可选，如 2024-01-01'
              />
              <Form.Input
                label='类型'
                name='pushme_type'
                onChange={handleInputChange}
                value={inputs.pushme_type}
                placeholder='可选，如 text'
              />
            </Form.Group>
          </>
        );
      case 'client':
        return (
          <>
            <Message>
              通过 WebSocket
              客户端进行推送，可以使用官方客户端实现，或者根据协议自行实现。官方客户端
              <a
                target='_blank'
                href='https://github.com/songquanpeng/personal-assistant'
              >
                详见此处
              </a>
              。
            </Message>
            <Form.Group widths='equal'>
              <Form.Input
                label='客户端连接密钥'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此设置客户端连接密钥'
              />
            </Form.Group>
          </>
        );
      case 'telegram':
        return (
          <>
            <Message>
              向
              <a href='https://t.me/botfather' target='_blank'>
                BotFather
              </a>
              创建机器人并填写令牌；先给机器人发任意一条消息，再点「获取会话
              ID」，最后保存。
              <br />
              API 地址、代理为可选项：留空使用官方 API；无法直连时可填 HTTP /
              SOCKS5 代理，例如{' '}
              <code>http://127.0.0.1:7890</code>、
              <code>socks5://127.0.0.1:7891</code>
              ；带账号密码时写成{' '}
              <code>http://user:pass@127.0.0.1:7890</code>、
              <code>socks5://user:pass@127.0.0.1:7891</code>。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='API 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='不填则使用官方地址'
              />
              <Form.Input
                label='代理地址'
                name='other'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.other}
                placeholder='可选，如 socks5://user:pass@127.0.0.1:7891'
              />
            </Form.Group>
            <Form.Group widths={2}>
              <Form.Input
                label='Telegram 机器人令牌'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此设置 Telegram 机器人令牌'
              />
              <Form.Input
                label='Telegram 会话 ID'
                name='account_id'
                type='text'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='在此设置 Telegram 会话 ID'
              />
            </Form.Group>
            <Button onClick={getTelegramChatId} loading={loading}>
              获取会话 ID
            </Button>
          </>
        );
      case 'discord':
        return (
          <>
            <Message>
              通过 Discord 群机器人进行推送，配置流程：选择一个 channel -&gt; 设置
              -&gt; 整合 -&gt; 创建 Webhook -&gt; 点击复制 Webhook URL。
              国内服务器通常需要填写代理才能访问 Discord。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='Webhook 地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写 Discord 提供的 Webhook 地址'
              />
              <Form.Input
                label='代理地址'
                name='other'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.other}
                placeholder='可选，如 http://127.0.0.1:7890 或 socks5://127.0.0.1:1080'
              />
            </Form.Group>
          </>
        );
      case 'one_bot':
        return (
          <>
            <Message>
              通过 OneBot 协议进行推送，可以使用{' '}
              <a href='https://github.com/Mrs4s/go-cqhttp' target='_blank'>
                cqhttp
              </a>{' '}
              等实现。 利用 OneBot 协议可以实现推送 QQ 消息。
              <br />
              注意，如果推送目标是群号则前面必须加上群号前缀，例如
              group_123456789。
            </Message>
            <Form.Group widths={3}>
              <Form.Input
                label='服务器地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写服务器地址'
              />
              <Form.Input
                label='推送 key'
                name='secret'
                type='password'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='在此填写服务器的 access token'
              />
              <Form.Input
                label='默认推送目标'
                name='account_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='在此填写默认推送目标，例如 QQ 号'
              />
            </Form.Group>
          </>
        );
      case 'group':
        return (
          <>
            <Message>
              对渠道进行分组，然后在推送时选择分组进行推送，可以实现一次性推送到多个渠道的功能。
              <br />
              <br />
              推送目标如若不填，则使用子渠道的默认推送目标。如果填写，请务必全部按顺序填写，对于不需要指定的直接留空即可，例如{' '}
              <code>123456789||@wechat</code>，两个连续的分隔符表示跳过该渠道。
            </Message>
            <Form.Group widths={2}>
              <Form.Input
                label='渠道列表'
                name='app_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.app_id}
                placeholder='在此填写渠道列表，使用 | 分割，例如 bark|telegram|wechat'
              />
              <Form.Input
                label='默认推送目标'
                name='account_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='在此填写默认推送目标，使用 | 分割，例如 123456789|@wechat|@wechat'
              />
            </Form.Group>
          </>
        );
      case 'lark_app':
        return (
          <>
            <Message>
              通过飞书自建应用进行推送，点击前往配置：
              <a target='_blank' href='https://open.feishu.cn/app'>
                飞书开放平台
              </a>
              。
              <br />
              需要为应用添加机器人能力：应用能力-&gt;添加应用能力—&gt;机器人。
              <br />
              需要为应用添加消息发送权限：开发配置-&gt;权限管理-&gt;权限配置-&gt;搜索「获取与发送单聊、群组消息」-&gt;开通权限。
              <br />
              注意，添加完成权限后需要发布版本提交审核才能见效。
              <br />
              注意，推送目标的格式为：
              <strong>
                <code>类型:ID</code>
              </strong>
              ，详见飞书
              <a
                href='https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create#bc6d1214'
                target='_blank'
              >
                开发文档
              </a>
              中查询参数一节。
            </Message>
            <Form.Group widths={3}>
              <Form.Input
                label='App ID'
                name='app_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.app_id}
                placeholder='应用凭证 -> App ID'
              />
              <Form.Input
                label='App Secret'
                name='secret'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.secret}
                placeholder='应用凭证 -> App Secret'
              />
              <Form.Input
                label='默认推送目标'
                name='account_id'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.account_id}
                placeholder='格式必须为：<类型>:<ID>，例如 open_id:123456'
              />
            </Form.Group>
          </>
        );
      case 'custom':
        return (
          <>
            <Message>
              自定义推送，目前仅支持 POST 请求，请求体为 JSON 格式。
              <br />
              支持以下模板变量：<code>$title</code>，<code>$description</code>，
              <code>$content</code>，<code>$url</code>，<code>$to</code>
              （请求体与请求头均可使用）。
              <br />
              <a
                href='https://iamazing.cn/page/message-pusher-common-custom-templates'
                target='_blank'
              >
                这个页面
              </a>
              给出了常见的第三方平台的配置实例，你可以参考这些示例进行配置。
              <br />
              注意，为了防止攻击者利用本功能访问内部网络，也为了你的信息安全，请求地址必须使用
              HTTPS 协议。
            </Message>
            <Form.Group widths='equal'>
              <Form.Input
                label='请求地址'
                name='url'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.url}
                placeholder='在此填写完整的请求地址，必须使用 HTTPS 协议'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.TextArea
                label='请求头'
                placeholder='可选，JSON 对象，例如 {"Authorization":"Bearer xxx"}'
                value={inputs.custom_headers}
                name='custom_headers'
                onChange={handleInputChange}
                style={{
                  minHeight: 100,
                  fontFamily: 'JetBrains Mono, Consolas',
                }}
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.TextArea
                label='请求体'
                placeholder='在此输入请求体，支持模板变量，必须为合法的 JSON 格式'
                value={inputs.other}
                name='other'
                onChange={handleInputChange}
                style={{
                  minHeight: 200,
                  fontFamily: 'JetBrains Mono, Consolas',
                }}
              />
            </Form.Group>
          </>
        );
      case 'none':
        return (
          <>
            <Message>
              仅保存消息，不做推送，可以在 Web
              端查看，需要用户具有消息持久化的权限。
            </Message>
          </>
        );
      default:
        return (
          <>
            <Message>未知通道类型！</Message>
          </>
        );
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>{isEditing ? '更新通道配置' : '新建消息通道'}</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={
                '请输入通道名称，请仅使用英文字母和下划线，该名称必须唯一'
              }
              onChange={handleInputChange}
              value={name}
              autoComplete='new-password'
              required
            />
          </Form.Field>
          <Form.Field>
            <Form.Input
              label='备注'
              name='description'
              type={'text'}
              placeholder={'请输入备注信息'}
              onChange={handleInputChange}
              value={description}
              autoComplete='new-password'
            />
          </Form.Field>
          <Form.Select
            label='通道类型'
            name='type'
            options={CHANNEL_OPTIONS}
            value={type}
            onChange={handleInputChange}
          />
          <Form.Input
            label='鉴权令牌'
            name='token'
            onChange={handleInputChange}
            autoComplete='new-password'
            value={inputs.token}
            placeholder='通道维度鉴权令牌，设置后使用该通道推送需要鉴权（使用全局鉴权令牌也可以）'
            action={{
              content: '随机生成',
              onClick: () => {
                setInputs((inputs) => ({
                  ...inputs,
                  token: generateToken(16),
                }));
              },
            }}
          />
          {renderChannelForm()}
          <Button as={Link} to='/channel'>
            返回
          </Button>
          <Button disabled={type === 'email'} onClick={submit}>
            提交
          </Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditChannel;

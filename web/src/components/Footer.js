import React, { useEffect, useState } from 'react';

import { Container, Segment } from 'semantic-ui-react';

const Footer = () => {
  const [footer, setFooter] = useState('');
  useEffect(() => {
    let savedFooter = localStorage.getItem('footer_html');
    if (!savedFooter) savedFooter = '';
    setFooter(savedFooter);
  });

  if (footer === '') {
    return null;
  }

  return (
    <Segment vertical>
      <Container textAlign='center'>
        <div
          className='custom-footer'
          dangerouslySetInnerHTML={{ __html: footer }}
        ></div>
      </Container>
    </Segment>
  );
};

export default Footer;

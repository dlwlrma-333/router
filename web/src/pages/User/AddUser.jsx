import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Card } from 'semantic-ui-react';
import { useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';

const AddUser = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const originInputs = {
    username: '',
    password: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const { username, password } = inputs;

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const submit = async () => {
    if (inputs.username === '' || inputs.password === '') return;
    if (password.length < 8 || password.length > 20) {
      showError(t('messages.error.password_length_range'));
      return;
    }
    const res = await API.post(`/api/v1/admin/user/`, inputs);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('user.messages.create_success'));
      setInputs(originInputs);
      navigate('/user');
    } else {
      showError(message);
    }
  };

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header router-page-title'>{t('user.add.title')}</Card.Header>
          <Form autoComplete='off'>
            <Form.Field>
              <Form.Input
                className='router-section-input'
                label={t('user.edit.username')}
                name='username'
                placeholder={t('user.edit.username_placeholder')}
                onChange={handleInputChange}
                value={username}
                autoComplete='off'
                required
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                className='router-section-input'
                label={t('user.edit.password')}
                name='password'
                type='password'
                placeholder={t('user.edit.password_placeholder')}
                onChange={handleInputChange}
                value={password}
                autoComplete='off'
                required
              />
            </Form.Field>
            <div className='router-toolbar-start router-block-gap-sm'>
              <Button className='router-page-button' positive type='submit' onClick={submit}>
                {t('user.edit.buttons.submit')}
              </Button>
            </div>
          </Form>
        </Card.Content>
      </Card>
    </div>
  );
};

export default AddUser;

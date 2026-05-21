import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Card,
  Statistic,
  Divider,
  Message,
  Icon,
} from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';
import { useTranslation } from 'react-i18next';

const TopUp = () => {
  const { t } = useTranslation();
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [user, setUser] = useState({});
  const [payMethods, setPayMethods] = useState([]);
  const [selectedMethod, setSelectedMethod] = useState('');
  const [amount, setAmount] = useState(1);
  const [minTopUp, setMinTopUp] = useState(1);

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo(t('topup.redeem_code.empty_code'));
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode,
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('topup.redeem_code.success'));
        setUserQuota((quota) => {
          return quota + data;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('topup.redeem_code.request_failed'));
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError(t('topup.redeem_code.no_link'));
      return;
    }
    let url = new URL(topUpLink);
    let username = user.username;
    let user_id = user.id;
    url.searchParams.append('username', username);
    url.searchParams.append('user_id', user_id);
    url.searchParams.append('transaction_id', crypto.randomUUID());
    window.open(url.toString(), '_blank');
  };

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setUserQuota(data.quota);
      setUser(data);
    } else {
      showError(message);
    }
  };

  const getTopUpInfo = async () => {
    try {
      let res = await API.get(`/api/user/topup/info`);
      const { success, data } = res.data;
      if (success) {
        setPayMethods(data.pay_methods || []);
        setMinTopUp(data.min_topup || 1);
        if (data.pay_methods && data.pay_methods.length > 0) {
          setSelectedMethod(data.pay_methods[0].type);
        }
      }
    } catch (err) {
      console.error('获取充值信息失败', err);
    }
  };

  const handleOnlinePay = async () => {
    if (!selectedMethod) {
      showError('请选择支付方式');
      return;
    }

    if (amount < minTopUp) {
      showError(`充值数量不能小于 ${minTopUp}`);
      return;
    }

    setIsSubmitting(true);
    try {
      let res;
      if (selectedMethod === 'stripe') {
        res = await API.post('/api/user/stripe/pay', {
          amount: amount,
          payment_method: 'stripe',
        });
      } else if (selectedMethod === 'alipay') {
        res = await API.post('/api/user/alipay/pay', {
          amount: amount,
          payment_method: 'alipay',
        });
      } else {
        showError('不支持的支付方式');
        return;
      }

      const { success, message, data } = res.data;
      if (success && data.pay_link) {
        window.location.href = data.pay_link;
      } else {
        showError(message || '支付创建失败');
      }
    } catch (err) {
      showError('支付请求失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    getUserQuota().then();
    getTopUpInfo().then();
  }, []);

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Header as='h2'>{t('topup.title')}</Header>
          </Card.Header>

          <Grid columns={2} stackable>
            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content
                  style={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <Card.Header>
                    <Header as='h3' style={{ color: '#2185d0', margin: '1em' }}>
                      <i className='credit card icon'></i>
                      在线充值
                    </Header>
                  </Card.Header>
                  <Card.Description
                    style={{
                      flex: 1,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      style={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'space-between',
                      }}
                    >
                      <div style={{ textAlign: 'center', paddingTop: '1em' }}>
                        <Statistic>
                          <Statistic.Value style={{ color: '#2185d0' }}>
                            {renderQuota(userQuota, t)}
                          </Statistic.Value>
                          <Statistic.Label>
                            {t('topup.get_code.current_quota')}
                          </Statistic.Label>
                        </Statistic>
                      </div>

                      {payMethods.length > 0 ? (
                        <div style={{ padding: '1em' }}>
                          <Form>
                            <Form.Field>
                              <label>支付方式</label>
                              <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                                {payMethods.map((method) => (
                                  <Button
                                    key={method.type}
                                    color={selectedMethod === method.type ? 'blue' : 'grey'}
                                    onClick={() => setSelectedMethod(method.type)}
                                    style={{
                                      backgroundColor: selectedMethod === method.type ? method.color : undefined,
                                    }}
                                  >
                                    {method.name}
                                  </Button>
                                ))}
                              </div>
                            </Form.Field>
                            <Form.Field>
                              <label>充值数量</label>
                              <Form.Input
                                type='number'
                                min={minTopUp}
                                value={amount}
                                onChange={(e) => setAmount(parseInt(e.target.value) || 0)}
                                placeholder={`最少 ${minTopUp}`}
                              />
                            </Form.Field>
                            <Button
                              primary
                              size='large'
                              fluid
                              onClick={handleOnlinePay}
                              loading={isSubmitting}
                              disabled={isSubmitting || !selectedMethod}
                            >
                              立即充值
                            </Button>
                          </Form>
                        </div>
                      ) : (
                        <Message info>
                          <Message.Header>暂无在线支付方式</Message.Header>
                          <p>请联系管理员开启在线支付功能</p>
                        </Message>
                      )}
                    </div>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>

            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content
                  style={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <Card.Header>
                    <Header as='h3' style={{ color: '#21ba45', margin: '1em' }}>
                      <i className='ticket alternate icon'></i>
                      {t('topup.redeem_code.title')}
                    </Header>
                  </Card.Header>
                  <Card.Description
                    style={{
                      flex: 1,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      style={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Form.Input
                        fluid
                        icon='key'
                        iconPosition='left'
                        placeholder={t('topup.redeem_code.placeholder')}
                        value={redemptionCode}
                        onChange={(e) => {
                          setRedemptionCode(e.target.value);
                        }}
                        onPaste={(e) => {
                          e.preventDefault();
                          const pastedText = e.clipboardData.getData('text');
                          setRedemptionCode(pastedText.trim());
                        }}
                        action={
                          <Button
                            icon='paste'
                            content={t('topup.redeem_code.paste')}
                            onClick={async () => {
                              try {
                                const text =
                                  await navigator.clipboard.readText();
                                setRedemptionCode(text.trim());
                              } catch (err) {
                                showError(t('topup.redeem_code.paste_error'));
                              }
                            }}
                          />
                        }
                      />

                      <div style={{ paddingBottom: '1em' }}>
                        <Button
                          color='green'
                          fluid
                          size='large'
                          onClick={topUp}
                          loading={isSubmitting}
                          disabled={isSubmitting}
                        >
                          {isSubmitting
                            ? t('topup.redeem_code.submitting')
                            : t('topup.redeem_code.submit')}
                        </Button>
                      </div>
                    </div>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid>
        </Card.Content>
      </Card>
    </div>
  );
};

export default TopUp;

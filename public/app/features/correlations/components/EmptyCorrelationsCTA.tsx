import React from 'react';

import { Card } from '@grafana/ui';
import { EmptyState } from '@grafana/ui/src/components/EmptyState/EmptyState';
import { EmptyStateCTAButton } from '@grafana/ui/src/components/EmptyState/EmptyStateCTAButton';
import { Trans, t } from 'app/core/internationalization';

interface Props {
  onClick: () => void;
  canWriteCorrelations: boolean;
}
export const EmptyCorrelationsCTA = ({ onClick, canWriteCorrelations }: Props) => {
  // TODO: if there are no datasources show a different message

  return canWriteCorrelations ? (
    <EmptyState
      button={
        <EmptyStateCTAButton
          buttonLabel={t('correlations.empty-state.button-title', 'Add correlation')}
          onButtonClick={onClick}
        />
      }
      message={t('correlations.empty-state.title', "You haven't defined any correlations yet")}
    >
      <Trans i18nKey="correlations.empty-state.pro-tip">
        You can also define correlations via datasource provisioning
      </Trans>
    </EmptyState>
  ) : (
    <Card>
      <Card.Heading>There are no correlations configured yet.</Card.Heading>
      <Card.Description>Please contact your administrator to create new correlations.</Card.Description>
    </Card>
  );
};

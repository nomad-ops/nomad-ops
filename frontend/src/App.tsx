import * as React from 'react';
import CssBaseline from '@mui/material/CssBaseline';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Link from '@mui/material/Link';
import { MainNavItems, SecondaryNavItems } from './nav/ListItems';

import { Key } from './domain/Key';
import RealTimeAccess from './services/RealTimeAccess';
import Nav from './components/Nav';
import { Outlet } from 'react-router-dom';
import { Source } from './domain/Source';

function Copyright(props: any) {
  return (
    <Typography variant="body2" color="text.secondary" align="center" {...props}>
      {'Copyright © '}
      <Link color="inherit" href="/">
        Nomad Ops
      </Link>{' '}
      {new Date().getFullYear()}
      {'.'}
    </Typography>
  );
}


function DashboardContent() {
  return (
    <Box sx={{ display: 'flex' }}>
      <CssBaseline />
      <Nav mainNavItems={MainNavItems} secondaryNavItems={SecondaryNavItems} />
      <Box
        component="main"
        sx={{
          backgroundColor: (theme) =>
            theme.palette.mode === 'light'
              ? theme.palette.grey[100]
              : theme.palette.grey[900],
          flexGrow: 1,
          height: '100vh',
          overflow: 'auto',
        }}
      >
        <Toolbar />
        <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
          <Outlet />
          <Copyright sx={{ pt: 4 }} />
        </Container>
      </Box>
    </Box>
  );
}

export default function App() {

  // Initialize stores ...
  RealTimeAccess.NewStore<Key>("keys", (record) => {
    var res: Key = {
      id: record.id,
      name: record["name"],
      value: record["value"],
      created: record.created
    };

    return res;
  });
  RealTimeAccess.NewStore<Source>("sources", (record) => {
    var res: Source = {
      id: record.id,
      name: record["name"],
      url: record["url"],
      path: record["path"],
      branch: record["branch"],
      dataCenter: record["dataCenter"],
      namespace: record["namespace"],
      region: record["region"],
      force: record["force"],
      created: record.created,
      updated: record.updated,
      status: record["status"],
      deployKey: record["deployKey"]
    };

    return res;
  });

  return <DashboardContent />;
}
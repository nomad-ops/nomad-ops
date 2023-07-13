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
import { useNavigate } from "react-router-dom";
import { Source } from './domain/Source';
import { Team } from './domain/Team';
import { User } from './domain/User';
import { VaultToken } from './domain/VaultToken';
import pb from './services/PocketBase';

function Copyright(props: any) {
  return (
    <Typography variant="body2" color="text.secondary" align="center" {...props}>
      {'Copyright © '}
      <Link color="inherit" href="/">
        Nomad Ops
      </Link>{' '}
      {"2022 - " + new Date().getFullYear()}
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
        <Container maxWidth={false}>
          <Outlet />
          <Copyright sx={{ pt: 4 }} />
        </Container>
      </Box>
    </Box>
  );
}

export default function App() {
  const navigate = useNavigate();
  pb.collection("users").authRefresh()
  .then(() => {

  }, () => {
    navigate("/login");
  });

  // Initialize stores ...
  RealTimeAccess.NewStore<User>("users", (record) => {
    var res: User = record as any;

    return res;
  });
  RealTimeAccess.NewStore<Key>("keys", (record) => {
    var res: Key = {
      id: record.id,
      name: record["name"],
      value: record["value"],
      team: record["team"],
      created: record.created
    };

    return res;
  });
  RealTimeAccess.NewStore<VaultToken>("vault_tokens", (record) => {
    var res: VaultToken = {
      id: record.id,
      name: record["name"],
      value: record["value"],
      team: record["team"],
      created: record.created
    };

    return res;
  });
  RealTimeAccess.NewStore<Team>("teams", (record) => {
    var res: Team = {
      id: record.id,
      name: record["name"],
      members: record["members"],
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
      paused: record["paused"],
      created: record.created,
      updated: record.updated,
      status: record["status"],
      deployKey: record["deployKey"],
      vaultToken: record["vaultToken"],
      teams: record["teams"]
    };

    return res;
  });

  return <DashboardContent />;
}
import React from "react";
import { Box, List, Drawer, Toolbar} from "@mui/material";
import { Source } from "../domain/Source";

import { JsonView, allExpanded, defaultStyles } from 'react-json-view-lite';
import 'react-json-view-lite/dist/index.css';

function removeEmpty(data: any): any {
    //transform properties into key-values pairs and filter all the empty-values
    const entries = Object.entries(data).filter(([, value]) => value != null);
  
    //map through all the remaining properties and check if the value is an object.
    //if value is object, use recursion to remove empty properties
    const clean = entries.map(([key, v]) => {
      const value = typeof v == 'object' ? removeEmpty(v) : v;
      return [key, value];
    });
  
    //transform the key-value pairs back to an object.
    return Object.fromEntries(clean);
}

// This function recursively sets all objects with a property "Type" with the value 'None' to undefined
function removeTypeNone(data: any): any {
    const entries = Object.entries(data).filter(([key, value]) => {
        if (key === 'Type' && value === 'None') {
            return undefined
        } else {
            return true
        }
    });
  
    const clean = entries.map(([key, v]) => {
      const value = typeof v == 'object' ? removeTypeNone(v) : v;

      return [key, value];
    });

    return Object.fromEntries(clean);
}

function removeUndefinedElements(data: any): any {
    // remove undefined elements in objects recursively
    if (Array.isArray(data)) {
        return data.filter((value) => value !== undefined).map((value) => removeUndefinedElements(value));
    } else if (typeof data === 'object') {
        return Object.fromEntries(Object.entries(data).filter(([key, value]) => value !== undefined).map(([key, value]) => [key, removeUndefinedElements(value)]));
    } else {
        return data;
    }
}
  

export default function SourceDiffDrawer({ open, onClose, source }: {
    open: boolean,
    onClose: (event: React.KeyboardEvent | React.MouseEvent) => void,
    source: Source
}) {
    React.useEffect(() => {
        if (source.status === undefined) {
            return;
        }
        if (source.status?.jobs === undefined) {
            return;
        }
    }, [source]);

    function cleanDiff(diff: any): any {
        return removeUndefinedElements(removeTypeNone(removeEmpty(diff)))
    }

    const list = () => (
        <Box
            sx={{ width: 900, padding: "2em" }}
            role="presentation"
        >
            <Box component="h3" sx={{margin: "0"}}>Diff View</Box>
            <List>
                { source.status?.jobs === undefined ? <div></div> : 
                Object.keys(source.status?.jobs).map((key) => {
                    let diff = cleanDiff(source.status!.jobs![key].diff)
                    return (
                    <div>
                        <h4>Job - {key}</h4>
                        <React.Fragment>
                            <JsonView data={diff} shouldExpandNode={allExpanded} style={defaultStyles} />
                        </React.Fragment>
                    </div>
                    )
                })}
            </List>
        </Box >
    );
    return <Drawer
        anchor={'right'}
        open={open}
        onClose={onClose}
    >
        <Toolbar />
        {list()}
    </Drawer>
}
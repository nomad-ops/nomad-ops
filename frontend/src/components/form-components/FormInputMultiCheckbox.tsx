import { FormControl, FormLabel, FormControlLabel, Checkbox, List, ListItem } from "@mui/material";
import React, { useEffect, useState } from "react";
import { Controller } from "react-hook-form";
import { FormInputProps } from "./FormInputProps";

export const FormInputMultiCheckbox: React.FC<FormInputProps> = ({
    name,
    control,
    setValue,
    label,
    options
}) => {
    const [selectedItems, setSelectedItems] = useState<any>(options?.filter((opt) => {
        return opt.selected === true;
    }).map((opt) => {
        return opt.value;
    }));

    const handleSelect = (value: any) => {
        const isPresent = selectedItems.indexOf(value);
        if (isPresent !== -1) {
            const remaining = selectedItems.filter((item: any) => item !== value);
            setSelectedItems(remaining);
        } else {
            setSelectedItems((prevItems: any) => [...prevItems, value]);
        }
    };

    useEffect(() => {
        setValue(name, selectedItems);
    }, [selectedItems, name, setValue]);

    return (
        <FormControl size={"small"} variant={"outlined"}>
            <FormLabel component="legend">{label}</FormLabel>

            <List dense={true} disablePadding={true}>
                {options?.map((option: any) => {
                    return (
                        <ListItem key={option.value}>
                            <FormControlLabel
                                control={
                                    <Controller
                                        name={name}
                                        render={() => {
                                            return (
                                                <Checkbox
                                                    checked={selectedItems.includes(option.value)}
                                                    onChange={() => handleSelect(option.value)}
                                                />
                                            );
                                        }}
                                        control={control}
                                    />
                                }
                                label={option.label}
                            />
                        </ListItem>
                    );
                })}
            </List>
        </FormControl>
    );
};
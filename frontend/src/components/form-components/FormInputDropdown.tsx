import { MenuItem, FormControl, InputLabel, Select } from "@mui/material";
import React from "react";
import { Controller } from "react-hook-form";
import { FormInputProps } from "./FormInputProps";

export const FormInputDropdown: React.FC<FormInputProps> = ({
    name,
    control,
    label,
    options,
    required
}) => {
    const generateSingleOptions = () => {
        return options?.map((option: any) => {
            return (
                <MenuItem key={option.value} value={option.value}>
                    {option.label}
                </MenuItem>
            );
        });
    };

    return (
        <FormControl fullWidth sx={{ marginTop: "1rem" }}>
            <InputLabel id={name}>{label}</InputLabel>
            <Controller
                render={({ field: { onChange, value } }) => (
                    <Select
                        labelId={name}
                        label={label}
                        name={name}
                        fullWidth
                        onChange={onChange} value={value} required={required}>
                        {generateSingleOptions()}
                    </Select>
                )}
                control={control}
                name={name}
            />
        </FormControl>
    );
};
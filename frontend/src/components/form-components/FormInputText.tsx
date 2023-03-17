import { TextField } from "@mui/material";
import { Controller } from "react-hook-form";
import { FormInputProps } from "./FormInputProps";

export const FormInputText = ({ name, control, label, autoFocus, type, required }: FormInputProps) => {
    return (
        <Controller
            name={name}
            control={control}
            render={({
                field: { onChange, value },
                fieldState: { error },
                formState,
            }) => (
                <TextField
                    helperText={error ? error.message : null}
                    autoFocus={autoFocus}
                    required={required}
                    size="small"
                    error={!!error}
                    onChange={onChange}
                    type={type}
                    value={value}
                    fullWidth
                    label={label}
                    variant="outlined"
                    margin="dense"
                />
            )}
        />
    );
};
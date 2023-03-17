import { Typography } from "@mui/material";
import { Controller } from "react-hook-form";
import { FormInputProps } from "./FormInputProps";

export const FormTextArea = ({ name, control, label, autoFocus, type, required }: FormInputProps) => {
    return (
        <Controller
            name={name}
            control={control}
            render={({
                field: { onChange, value },
                fieldState: { error },
                formState,
            }) => (
                <div>
                    <Typography variant="body2" color="text.secondary">{label}</Typography>
                    <textarea
                        autoFocus={autoFocus}
                        required={required}
                        onChange={onChange}
                        value={value}
                        style={{ width: "100%", resize: "none", height: "300px" }}

                    />
                </div>
            )}
        />
    );
};
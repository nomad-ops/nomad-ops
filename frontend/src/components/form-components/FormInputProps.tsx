export interface FormInputProps {
    name: string;
    type?: string;
    control: any;
    label: string;
    required?: boolean,
    autoFocus?: boolean,
    setValue?: any;
    options?: { label: string, value: string }[]
}
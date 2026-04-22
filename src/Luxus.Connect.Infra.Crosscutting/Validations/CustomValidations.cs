using System.Text.RegularExpressions;

namespace Luxus.Connect.Infra.Crosscutting.Validations;

public static partial class CustomValidations
{
    [GeneratedRegex("[^0-9]")]
    private static partial Regex OnlyDigitsRegex();

    public static bool IsValidCnpj(string cnpj)
    {
        int[] multiplier1 = [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2];
        int[] multiplier2 = [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2];

        string tempCnpj, digit;
        int sum, rest;

        string formatedCnpj = OnlyDigitsRegex().Replace(cnpj, "");

        if (formatedCnpj.Length != 14 || formatedCnpj.Distinct().Count() == 1)
        {
            return false;
        }

        tempCnpj = formatedCnpj[..12];
        sum = 0;

        for (int i = 0; i < 12; i++)
        {
            sum += int.Parse(tempCnpj[i].ToString()) * multiplier1[i];
        }

        rest = sum % 11;
        rest = rest < 2 ? 0 : 11 - rest;

        digit = rest.ToString();
        tempCnpj += digit;
        sum = 0;

        for (int i = 0; i < 13; i++)
        {
            sum += int.Parse(tempCnpj[i].ToString()) * multiplier2[i];
        }

        rest = sum % 11;
        rest = rest < 2 ? 0 : 11 - rest;

        digit += rest.ToString();
        return formatedCnpj.EndsWith(digit);
    }

    public static bool IsValidCpf(string cpf)
    {
        int[] multiplier1 = [10, 9, 8, 7, 6, 5, 4, 3, 2];
        int[] multiplier2 = [11, 10, 9, 8, 7, 6, 5, 4, 3, 2];

        string tempCpf, digit;
        int sum, rest;

        string formatedCpf = OnlyDigitsRegex().Replace(cpf, "");

        if (formatedCpf.Length != 11 || formatedCpf.Distinct().Count() == 1)
        {
            return false;
        }

        tempCpf = formatedCpf[..9];
        sum = 0;

        for (int i = 0; i < 9; i++)
        {
            sum += int.Parse(tempCpf[i].ToString()) * multiplier1[i];
        }

        rest = sum % 11;
        rest = rest < 2 ? 0 : 11 - rest;

        digit = rest.ToString();
        tempCpf += digit;
        sum = 0;

        for (int i = 0; i < 10; i++)
        {
            sum += int.Parse(tempCpf[i].ToString()) * multiplier2[i];
        }

        rest = sum % 11;

        rest = rest < 2 ? 0 : 11 - rest;
        digit += rest.ToString();

        return formatedCpf.EndsWith(digit);
    }
}

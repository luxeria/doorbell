# this script implements a small awk-based template engine which replaces 
# {{ VARNAME }} with the contents of the corresponding environment variable
{
  # repeats until no matches in current line anymore
  while (match($0, /\{\{\s*\w+\s*\}\}/)) {
    # extract matched substring
    m = substr($0, RSTART, RLENGTH)
    # remove surrounding braces
    var = substr(m, 3, length(m)-4)
    # trim whitespace from var
    gsub(/\s/, "", var)
    # replace matched substring with the env variable
    sub(m, ENVIRON[var], $0)
  }
  # print processed line
  print
}

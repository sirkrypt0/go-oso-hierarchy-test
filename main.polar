allow(actor, action, resource) if
  has_permission(actor, action, resource);

actor User {}

# Teams
resource Team {
    permissions = ["read", "edit", "admin"];
    roles = ["guest", "researcher", "maintainer", "owner"];

    relations = { parent: Team };

    "read" if "guest";
    "edit" if "maintainer";
    "admin" if "owner";

    "guest" if "researcher";
    "researcher" if "maintainer";
    "maintainer" if "owner";

    # Check hierarchy
    "guest" if "guest" on "parent";
    "researcher" if "researcher" on "parent";
    "maintainer" if "maintainer" on "parent";
    "owner" if "owner" on "parent";
}

has_relation(parent: Team, "parent", team: Team) if
    # Apparently, we need this to stop Oso from going higher up the hierarchy.
    # Otherwise, it would expect that it can find a team with ID 0 and fail ... 
    team.ParentID != 0 and
    team.Parent = parent;

has_role(user: User, name: String, team: Team) if
    role in user.Teams and
    role.Role = name and
    role.TeamID = team.ID;

# Team Resource
resource Repository {
    permissions = ["read", "edit", "admin"];
    relations = { team: Team };

    "read" if "guest" on "team";
    "edit" if "maintainer" on "team";
    "admin" if "owner" on "team";
}

has_relation(team: Team, "team", repo: Repository) if
    team = repo.Team;
